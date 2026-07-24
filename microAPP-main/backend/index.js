// Load environment variables
require('dotenv').config();

const express = require('express');
const bodyParser = require('body-parser');
const cors = require('cors');
const fs = require('fs');
const path = require('path');
const XLSX = require('xlsx');
const multer = require('multer');
const jwt = require('jsonwebtoken');
const bcrypt = require('bcryptjs');
const { randomUUID } = require('crypto');
const axios = require('axios');
const { queryMemberByAuthCode } = require('./platformAuth');
const { config, validateConfig, printConfig } = require('./config');
const { logger } = require('./logger');
const { sanitizeBody } = require('./middleware/validation');
const { errorHandler, notFoundHandler, asyncHandler } = require('./middleware/error');
const adminRouter = require('./routes/admin');
const { router: medicalRouter, startTimeoutChecker } = require('./routes/medical');

const app = express();
const PORT = config.server.port;

// 验证配置
const configValidation = validateConfig();
if (configValidation.warnings.length > 0) {
    configValidation.warnings.forEach(warning => logger.warn(warning));
}

const {
    initStorage,
    readUsersDB,
    writeUsersDB,
    readCasesDB,
    writeCasesDB,
    readCaseCategoriesDB,
    writeCaseCategoriesDB,
    readApplicationsDB,
    writeApplicationsDB,
    readServicesConfig,
    writeServicesConfig,
    readReviewsDB,
    writeReviewsDB
} = require('./storage');

// 使用配置模块中的值
const JWT_SECRET = config.jwt.secret;
const ACCESS_TOKEN_TTL = config.jwt.accessTokenTTL;
const REFRESH_TOKEN_TTL = config.jwt.refreshTokenTTL;
const SUPER_ADMIN_PHONE = config.admin.superAdminPhone;
const WX_APPID = config.wechat.appId;
const WX_SECRET = config.wechat.appSecret;
const WX_MP_APPID = config.wechat.mpAppId;
const WX_MP_SECRET = config.wechat.mpAppSecret;

let wxAccessTokenCache = { token: '', expiresAt: 0 };

async function getWxAccessToken() {
    if (wxAccessTokenCache.token && Date.now() < wxAccessTokenCache.expiresAt - 60000) {
        return wxAccessTokenCache.token;
    }
    const url = `https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=${WX_APPID}&secret=${WX_SECRET}`;
    const { data } = await axios.get(url);
    if (data.errcode) {
        throw new Error(`获取微信access_token失败: ${data.errmsg}`);
    }
    wxAccessTokenCache = {
        token: data.access_token,
        expiresAt: Date.now() + data.expires_in * 1000
    };
    return data.access_token;
}

// CORS 配置
app.use(cors({
    origin: config.server.corsOrigin,
    credentials: true
}));

app.use(bodyParser.json({ limit: '50mb' }));
app.use(bodyParser.urlencoded({ limit: '50mb', extended: true }));

// 输入消毒
app.use(sanitizeBody);

// Serve static files with cache headers
app.use(express.static(path.join(__dirname, 'public'), {
    maxAge: '7d',
    etag: true,
    lastModified: true,
    immutable: false
}));

// Image compression endpoint with disk cache + WebP support
const sharp = require('sharp');
const IMAGE_CACHE_DIR = path.join(__dirname, '.image-cache');
if (!fs.existsSync(IMAGE_CACHE_DIR)) {
    fs.mkdirSync(IMAGE_CACHE_DIR, { recursive: true });
}

app.get('/api/image', asyncHandler(async (req, res) => {
    const { url, width, quality, format: formatParam } = req.query;
    const imagePath = url ? path.join(__dirname, 'public', url.replace(/^\//, '')) : null;

    if (!imagePath || !fs.existsSync(imagePath)) {
        return res.status(404).json({ error: 'Image not found' });
    }

    const imgWidth = parseInt(width) || 800;
    const imgQuality = parseInt(quality) || 75;

    // Determine output format: prefer WebP if client supports it
    const acceptHeader = req.headers['accept'] || '';
    const supportsWebP = acceptHeader.includes('image/webp');
    const outputFormat = formatParam || (supportsWebP ? 'webp' : 'jpeg');
    const ext = outputFormat === 'webp' ? 'webp' : 'jpg';
    const contentType = outputFormat === 'webp' ? 'image/webp' : 'image/jpeg';

    // Generate cache key from source file mtime + params
    const srcStat = fs.statSync(imagePath);
    const cacheKey = `${path.basename(url, path.extname(url))}_${imgWidth}_${imgQuality}_${ext}_${srcStat.mtimeMs}`;
    const cachePath = path.join(IMAGE_CACHE_DIR, `${cacheKey}.${ext}`);

    // Serve from disk cache if available
    if (fs.existsSync(cachePath)) {
        res.set('Content-Type', contentType);
        res.set('Cache-Control', 'public, max-age=2592000, immutable'); // 30 days
        res.set('Vary', 'Accept');
        return res.sendFile(cachePath);
    }

    try {
        let pipeline = sharp(imagePath)
            .resize(imgWidth, null, { fit: 'inside', withoutEnlargement: true });

        if (outputFormat === 'webp') {
            pipeline = pipeline.webp({ quality: imgQuality });
        } else {
            pipeline = pipeline.jpeg({ quality: imgQuality, progressive: true });
        }

        // Write to disk cache, then serve
        await pipeline.toFile(cachePath);

        res.set('Content-Type', contentType);
        res.set('Cache-Control', 'public, max-age=2592000, immutable'); // 30 days
        res.set('Vary', 'Accept');
        res.sendFile(cachePath);
    } catch (err) {
        // Fallback: serve original image
        logger.warn('Image compression failed, serving original:', err.message);
        res.set('Content-Type', 'image/jpeg');
        res.set('Cache-Control', 'public, max-age=3600');
        res.sendFile(imagePath);
    }
}));

// 请求日志
app.use((req, res, next) => {
    const start = Date.now();
    res.on('finish', () => {
        const duration = Date.now() - start;
        logger.info(`${req.method} ${req.originalUrl}`, {
            status: res.statusCode,
            duration: `${duration}ms`,
            ip: req.ip
        });
    });
    next();
});

function getClientIp(req) {
    const xf = req.headers['x-forwarded-for'];
    let ip =
        (typeof xf === 'string' && xf.split(',')[0].trim()) ||
        (Array.isArray(xf) && xf[0]) ||
        req.socket?.remoteAddress ||
        req.ip ||
        '';
    if (ip.startsWith('::ffff:')) ip = ip.slice(7);
    return ip;
}

function hashToSeed(str) {
    // FNV-1a 32-bit
    let h = 2166136261;
    const s = String(str || '');
    for (let i = 0; i < s.length; i++) {
        h ^= s.charCodeAt(i);
        h = Math.imul(h, 16777619);
    }
    return h >>> 0;
}

function sanitizeUser(user) {
    if (!user) return null;
    const { password, passwordHash, refreshToken, refreshTokenExpiresAt, wxSessionKey, ...safe } = user;
    return safe;
}

function generateTokens(user) {
    const accessToken = jwt.sign(
        { sub: user.id, role: user.role },
        JWT_SECRET,
        { expiresIn: ACCESS_TOKEN_TTL }
    );
    const refreshToken = jwt.sign(
        { sub: user.id, type: 'refresh' },
        JWT_SECRET,
        { expiresIn: REFRESH_TOKEN_TTL }
    );
    const decoded = jwt.decode(refreshToken) || {};
    return {
        accessToken,
        refreshToken,
        refreshTokenExpiresAt: decoded.exp ? decoded.exp * 1000 : null
    };
}

async function findUserById(userId) {
    const users = await readUsersDB();
    return users.find(u => u.id === userId);
}

async function authRequired(req, res, next) {
    const authHeader = req.headers.authorization || '';
    const token = authHeader.startsWith('Bearer ') ? authHeader.slice(7) : '';
    if (!token) {
        return res.status(401).json({ success: false, message: '未登录或登录已过期' });
    }
    try {
        const payload = jwt.verify(token, JWT_SECRET);
        const user = await findUserById(payload.sub);
        if (!user) {
            return res.status(401).json({ success: false, message: '用户不存在或登录已失效' });
        }
        req.user = sanitizeUser(user);
        next();
    } catch (err) {
        return res.status(401).json({ success: false, message: '登录已过期，请重新登录' });
    }
}

async function authOptional(req, res, next) {
    const authHeader = req.headers.authorization || '';
    const token = authHeader.startsWith('Bearer ') ? authHeader.slice(7) : '';
    if (!token) {
        return next();
    }
    try {
        const payload = jwt.verify(token, JWT_SECRET);
        const user = await findUserById(payload.sub);
        if (user) {
            req.user = sanitizeUser(user);
        }
    } catch (err) {
        // ignore invalid token for optional auth
    }
    next();
}

function roleRequired(roles = []) {
    return (req, res, next) => {
        const user = req.user;
        if (!user) {
            return res.status(401).json({ success: false, message: '未登录或登录已过期' });
        }
        if (!roles.includes(user.role)) {
            return res.status(403).json({ success: false, message: '无权限访问' });
        }
        next();
    };
}

async function verifyUserPassword(user, inputPassword) {
    if (!user || !inputPassword) return false;
    if (user.passwordHash) {
        return bcrypt.compare(inputPassword, user.passwordHash);
    }
    return user.password === inputPassword;
}

// Client IP helper (LAN testing / 分发)
app.get('/api/client-ip', (req, res) => {
    const ip = getClientIp(req);
    res.json({ success: true, ip });
});

// Game assignment based on IP (seed + bucket for future sharding)
app.get('/api/games/assign', (req, res) => {
    const ip = getClientIp(req);
    const seed = hashToSeed(ip || 'unknown');
    res.json({
        success: true,
        ip,
        playerKey: ip || 'unknown',
        seed,
        bucket: seed % 4,
        assignedAt: new Date().toISOString()
    });
});

// Multer Storage Configuration
const storage = multer.diskStorage({
    destination: function (req, file, cb) {
        const uploadDir = path.join(__dirname, 'public/uploads');
        if (!fs.existsSync(uploadDir)){
            fs.mkdirSync(uploadDir, { recursive: true });
        }
        cb(null, uploadDir);
    },
    filename: function (req, file, cb) {
        const uniqueSuffix = Date.now() + '-' + Math.round(Math.random() * 1E9);
        cb(null, uniqueSuffix + path.extname(file.originalname));
    }
});

const upload = multer({ storage: storage });

// Upload Endpoint
app.post('/api/upload', upload.single('file'), (req, res) => {
    if (!req.file) {
        return res.status(400).json({ success: false, message: 'No file uploaded' });
    }
    const fileUrl = '/uploads/' + req.file.filename;
    res.json({ success: true, url: fileUrl });
});

// Storage init happens in bootstrap before server starts

// Ensure DSL admin and standard admin exist (Patch)
async function ensureAdminUsers() {
    let users = await readUsersDB();
    let modified = false;

    const superAdminIndex = users.findIndex(u => u.phone === SUPER_ADMIN_PHONE);
    if (superAdminIndex === -1) {
        users.push({
            id: randomUUID(),
            phone: SUPER_ADMIN_PHONE,
            password: 'admin',
            passwordHash: bcrypt.hashSync('admin', 10),
            name: SUPER_ADMIN_PHONE,
            role: 'admin',
            avatar: '',
            createTime: new Date().toISOString()
        });
        modified = true;
    } else if (users[superAdminIndex].role !== 'admin') {
        users[superAdminIndex].role = 'admin';
        modified = true;
    }

    // Check DSLadmin
    if (!users.find(u => u.phone === 'DSLadmin')) {
        users.push({
            id: 'dsl_admin_id',
            phone: 'DSLadmin',
            password: 'dkjjfwy2026DSL', // legacy plain text (kept for compatibility)
            passwordHash: bcrypt.hashSync('dkjjfwy2026DSL', 10),
            name: 'DSL管理员',
            role: 'dsl_admin',
            avatar: '',
            createTime: new Date().toISOString()
        });
        modified = true;
    }

    // Check study_admin (研学管理员)
    const studyAdminIndex = users.findIndex(u => u.phone === 'studyadmin');
    if (studyAdminIndex === -1) {
        users.push({
            id: 'study_admin_id',
            phone: 'studyadmin',
            password: 'dkyxAdmin',
            passwordHash: bcrypt.hashSync('dkyxAdmin', 10),
            name: '研学管理员',
            role: 'study_admin',
            avatar: '',
            createTime: new Date().toISOString()
        });
        modified = true;
    } else {
        // 确保密码已更新为最新
        const sa = users[studyAdminIndex];
        if (!sa.passwordHash || !bcrypt.compareSync('dkyxAdmin', sa.passwordHash)) {
            sa.password = 'dkyxAdmin';
            sa.passwordHash = bcrypt.hashSync('dkyxAdmin', 10);
            sa.role = 'study_admin';
            modified = true;
        }
    }

    // Ensure existing admins have passwordHash
    users = users.map(user => {
        if (user.role && !user.passwordHash && user.password) {
            modified = true;
            return { ...user, passwordHash: bcrypt.hashSync(user.password, 10) };
        }
        return user;
    });

    if (modified) {
        await writeUsersDB(users);
        console.log('Admin users patched/ensured.');
    }
}

async function ensureDefaultCases() {
    const existingCases = await readCasesDB();
    if (Array.isArray(existingCases) && existingCases.length) return;

    const defaultCases = [
        {
            id: 1764231459722,
            categoryId: 5,
            title: '泰顺无人机表演',
            description: '5月20日晚，泰顺县文祥湖公园的夜空被1999架无人机点亮，一场以“欢顺520”为主题的无人机灯光秀在此震撼上演。这场由温州交运集团所属低空公司、浙江顺翼泰翔低空经济有限公司联合泰顺县文化和广电旅游体育局共同打造的视觉盛宴，以天为幕、以光为笔，在高空书写了一封跨越半个多世纪的"时光情书"，用动态光影演绎"执子之手，与子偕老"的深情承诺，为这座浙南山城增添了一抹动人心魄的浪漫注脚。',
            service: '航空表演',
            location: '泰顺县文祥湖公园',
            date: '2025-05-20',
            coverType: 'image',
            cover: '/uploads/1764296713631-620487281.jpg',
            media: [
                { type: 'video', url: '/uploads/1764296721313-629351170.mp4' }
            ],
            fullDescription: '',
            highlights: []
        },
        {
            id: 1,
            categoryId: 1,
            title: '江心屿无人机外卖配送',
            description: '江心屿无人机外卖航线，极速送达，空投德克士、永和豆浆等多家外卖',
            service: '无人机物流服务',
            location: '江心屿',
            date: '2025-10-01',
            views: '1.2k',
            coverType: 'image',
            cover: '/uploads/1764296738238-439008154.jpg',
            media: [
                { type: 'image', url: '/uploads/1764296744506-267079065.jpg' },
                { type: 'image', url: '/uploads/1764296748498-687999548.jpg' },
                { type: 'video', url: '/uploads/1764296752919-602156690.mp4', poster: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?w=800' }
            ],
            fullDescription: '为缓解假期游客密集导致的就餐难等问题，温州交运集团所属低空公司在江心屿创新推出无人机外卖配送服务，以“宋街起飞、九曲桥草坪降落”的双节点布局，在景区内打造“低空物流”内循环。“相比传统地面配送，无人机配送效率提升了近70%，单次能装下2—3份餐食，有效解决配送慢、取餐难等问题。”低空公司一位负责人说，此次在江心屿推出无人机外卖配送服务，进一步探索了“低空物流+景区内循环”的可复制场景，为景区服务增添科技温度。',
            highlights: [
                '10分钟极速送达，效率提升300%',
                '精准定位空投柜，取餐更方便',
                '覆盖永和豆浆、德克士等种商品',
                '无接触配送，科技感十足'
            ]
        },
        {
            id: 4,
            categoryId: 4,
            title: '山区杨梅无人机吊运',
            description: '偏远山区农产品无人机空运，解决传统运输方式运输难题',
            service: '无人机吊运服务',
            location: '泰顺山区',
            date: '2025-06-24',
            views: '1.8k',
            coverType: 'image',
            cover: '/uploads/1764296782266-332050144.jpg',
            media: [
                { type: 'image', url: '/uploads/1764296786961-152168144.jpg' },
                { type: 'image', url: '/uploads/1764296790376-159415407.jpg' },
                { type: 'video', url: '/uploads/1764296794723-372267165.mp4' }
            ],
            fullDescription: '偏远山区，道路崎岖，传统载具运输困难。使用大型吊运无人机，成功杨梅进行吊装配送',
            highlights: [
                '突破地形限制，吊运能力强',
                '精准定位，误差小于10cm',
                '降低成本40%，缩短工期60%',
                '零安全事故，施工人员零风险'
            ]
        }
    ];

    await writeCasesDB(defaultCases);
}

// Applications DB wrappers (backed by JSON or PostgreSQL)
async function readDB() {
    return readApplicationsDB();
}

async function writeDB(data) {
    return writeApplicationsDB(data);
}

async function handleLogin(req, res) {
    try {
        const { phone, username, password } = req.body || {};
        const loginId = (phone || username || '').trim();
        if (!loginId || !password) {
            return res.status(400).json({ success: false, message: '账号或密码不能为空' });
        }
        const users = await readUsersDB();
        const user = users.find(u => u.phone === loginId || u.username === loginId);
        const isValid = await verifyUserPassword(user, password);
        if (!user || !isValid) {
            return res.status(401).json({ success: false, message: '账号或密码错误' });
        }

        // Upgrade legacy users to hashed password
        if (!user.passwordHash) {
            user.passwordHash = await bcrypt.hash(password, 10);
            user.password = user.password || '';
            await writeUsersDB(users);
        }

        const tokens = generateTokens(user);
        const updatedUsers = users.map(u => {
            if (u.id === user.id) {
                return { ...u, refreshToken: tokens.refreshToken, refreshTokenExpiresAt: tokens.refreshTokenExpiresAt };
            }
            return u;
        });
        await writeUsersDB(updatedUsers);

        res.json({ success: true, user: sanitizeUser(user), ...tokens });
    } catch (err) {
        console.error('[login] failed:', err);
        res.status(500).json({ success: false, message: '登录服务异常', detail: err?.message || 'unknown' });
    }
}

// Login Endpoint (legacy)
app.post('/api/login', handleLogin);

// Auth Login Endpoint
app.post('/api/auth/login', handleLogin);

// SSO Entry (redirect to H5 with authcode)
app.get('/sso/login', (req, res) => {
    const { authcode, jyauthcode, redirect } = req.query || {};
    const code = (authcode || jyauthcode || '').toString().trim();
    if (!code) {
        return res.status(400).send('缺少授权码');
    }
    const target = typeof redirect === 'string' && redirect.trim()
        ? redirect.trim()
        : '/#/home';
    const encodedAuth = encodeURIComponent(code);
    const separator = target.includes('?') ? '&' : '?';
    res.redirect(302, `${target}${separator}authcode=${encodedAuth}`);
});

// SSO Login (authcode from 畅行温州)
app.post('/api/sso/login', async (req, res) => {
    try {
        const { authcode, jyauthcode } = req.body || {};
        const code = (authcode || jyauthcode || '').toString().trim();
        if (!code) {
            return res.status(400).json({ success: false, message: '缺少授权码' });
        }

        const member = await queryMemberByAuthCode(code);
        const platformMemberNo = member.pltmemberno || member.authmemberno || member.mobilephone;
        if (!platformMemberNo) {
            return res.status(400).json({ success: false, message: '平台会员信息缺失' });
        }

        const users = await readUsersDB();
        let user = users.find(u => u.platformMemberNo === platformMemberNo);
        if (!user && member.mobilephone) {
            user = users.find(u => u.phone === member.mobilephone);
        }

        if (!user) {
            user = {
                id: randomUUID(),
                phone: member.mobilephone || '',
                password: '',
                passwordHash: '',
                name: member.nickname || member.name || `User${String(platformMemberNo).slice(-4)}`,
                role: 'user',
                avatar: '',
                platformMemberNo,
                platformSource: '畅行温州',
                createTime: new Date().toISOString()
            };
            users.push(user);
        } else {
            user.platformMemberNo = user.platformMemberNo || platformMemberNo;
            if (!user.phone && member.mobilephone) user.phone = member.mobilephone;
            if (!user.name && (member.nickname || member.name)) {
                user.name = member.nickname || member.name;
            }
        }

        const tokens = generateTokens(user);
        const updatedUsers = users.map(u => {
            if (u.id === user.id) {
                return { ...u, refreshToken: tokens.refreshToken, refreshTokenExpiresAt: tokens.refreshTokenExpiresAt };
            }
            return u;
        });
        await writeUsersDB(updatedUsers);

        res.json({ success: true, user: sanitizeUser(user), ...tokens });
    } catch (err) {
        console.error('[sso login] failed:', err);
        res.status(500).json({ success: false, message: err?.message || 'SSO登录失败' });
    }
});

// SSO Verify (no user creation)
app.post('/api/sso/verify', async (req, res) => {
    try {
        const { authcode, jyauthcode } = req.body || {};
        const code = (authcode || jyauthcode || '').toString().trim();
        if (!code) {
            return res.status(400).json({ success: false, message: '缺少授权码' });
        }
        const member = await queryMemberByAuthCode(code);
        res.json({
            success: true,
            member,
            platformMemberNo: member.pltmemberno || member.authmemberno || member.mobilephone || ''
        });
    } catch (err) {
        console.error('[sso verify] failed:', err);
        res.status(500).json({ success: false, message: err?.message || 'SSO验证失败' });
    }
});

// Register Endpoint (legacy)
app.post('/api/register', async (req, res) => {
    const { phone, password, name } = req.body || {};
    const users = await readUsersDB();
    if (!phone || !password) {
        return res.status(400).json({ success: false, message: '手机号或密码不能为空' });
    }
    if (users.find(u => u.phone === phone)) {
        return res.status(400).json({ success: false, message: '用户已存在' });
    }

    const newUser = {
        id: randomUUID(),
        phone,
        password: '', // do not keep plain text
        passwordHash: await bcrypt.hash(password, 10),
        name: name || `User${phone.slice(-4)}`,
        role: 'user',
        avatar: '',
        createTime: new Date().toISOString()
    };

    users.push(newUser);
    if (await writeUsersDB(users)) {
        const tokens = generateTokens(newUser);
        const updatedUsers = users.map(u => {
            if (u.id === newUser.id) {
                return { ...u, refreshToken: tokens.refreshToken, refreshTokenExpiresAt: tokens.refreshTokenExpiresAt };
            }
            return u;
        });
        await writeUsersDB(updatedUsers);
        res.json({ success: true, user: sanitizeUser(newUser), ...tokens });
    } else {
        res.status(500).json({ success: false, message: 'Failed to create user' });
    }
});

// Auth Register Endpoint
app.post('/api/auth/register', async (req, res) => {
    const { phone, password, name } = req.body || {};
    const users = await readUsersDB();
    if (!phone || !password) {
        return res.status(400).json({ success: false, message: '手机号或密码不能为空' });
    }
    if (users.find(u => u.phone === phone)) {
        return res.status(400).json({ success: false, message: '用户已存在' });
    }

    const newUser = {
        id: randomUUID(),
        phone,
        password: '',
        passwordHash: await bcrypt.hash(password, 10),
        name: name || `User${phone.slice(-4)}`,
        role: 'user',
        avatar: '',
        createTime: new Date().toISOString()
    };

    users.push(newUser);
    if (await writeUsersDB(users)) {
        const tokens = generateTokens(newUser);
        const updatedUsers = users.map(u => {
            if (u.id === newUser.id) {
                return { ...u, refreshToken: tokens.refreshToken, refreshTokenExpiresAt: tokens.refreshTokenExpiresAt };
            }
            return u;
        });
        await writeUsersDB(updatedUsers);
        res.json({ success: true, user: sanitizeUser(newUser), ...tokens });
    } else {
        res.status(500).json({ success: false, message: 'Failed to create user' });
    }
});

// Auth Me
app.get('/api/auth/me', authRequired, (req, res) => {
    res.json({ success: true, user: req.user });
});

// Auth Refresh
app.post('/api/auth/refresh', async (req, res) => {
    const { refreshToken } = req.body || {};
    if (!refreshToken) {
        return res.status(400).json({ success: false, message: '缺少 refresh token' });
    }
    let payload;
    try {
        payload = jwt.verify(refreshToken, JWT_SECRET);
    } catch (err) {
        return res.status(401).json({ success: false, message: 'refresh token 无效或已过期' });
    }
    if (payload.type !== 'refresh') {
        return res.status(401).json({ success: false, message: 'refresh token 类型不正确' });
    }

    const users = await readUsersDB();
    const user = users.find(u => u.id === payload.sub);
    if (!user || user.refreshToken !== refreshToken) {
        return res.status(401).json({ success: false, message: 'refresh token 无效' });
    }

    const accessToken = jwt.sign(
        { sub: user.id, role: user.role },
        JWT_SECRET,
        { expiresIn: ACCESS_TOKEN_TTL }
    );
    res.json({ success: true, accessToken });
});

// Auth Logout
app.post('/api/auth/logout', authRequired, async (req, res) => {
    const users = await readUsersDB();
    const updatedUsers = users.map(u => {
        if (u.id === req.user.id) {
            return { ...u, refreshToken: '', refreshTokenExpiresAt: null };
        }
        return u;
    });
    await writeUsersDB(updatedUsers);
    res.json({ success: true });
});

// WeChat Mini-Program Login (code → openid → find/create user → tokens)
app.post('/api/auth/wx-login', async (req, res) => {
    try {
        const { code } = req.body || {};
        if (!code) {
            return res.status(400).json({ success: false, message: '缺少微信授权码' });
        }
        if (!WX_APPID || !WX_SECRET) {
            return res.status(500).json({ success: false, message: '微信小程序配置缺失，请联系管理员' });
        }

        const wxUrl = `https://api.weixin.qq.com/sns/jscode2session?appid=${WX_APPID}&secret=${WX_SECRET}&js_code=${code}&grant_type=authorization_code`;
        const { data: wxRes } = await axios.get(wxUrl);

        if (wxRes.errcode) {
            console.error('[wx-login] code2Session failed:', wxRes);
            return res.status(400).json({ success: false, message: `微信授权失败: ${wxRes.errmsg}` });
        }

        const { openid, unionid, session_key } = wxRes;
        if (!openid) {
            return res.status(400).json({ success: false, message: '获取微信openid失败' });
        }

        const users = await readUsersDB();
        let user = users.find(u => u.wxOpenid === openid);
        let isNewUser = false;

        if (!user && unionid) {
            user = users.find(u => u.wxUnionid === unionid);
        }

        if (!user) {
            isNewUser = true;
            user = {
                id: randomUUID(),
                phone: '',
                password: '',
                passwordHash: '',
                name: `微信用户${openid.slice(-4)}`,
                role: 'user',
                avatar: '',
                wxOpenid: openid,
                wxUnionid: unionid || '',
                wxSessionKey: session_key || '',
                createTime: new Date().toISOString()
            };
            users.push(user);
        } else {
            if (unionid && !user.wxUnionid) user.wxUnionid = unionid;
            user.wxSessionKey = session_key || user.wxSessionKey || '';
        }

        const tokens = generateTokens(user);
        const updatedUsers = users.map(u => {
            if (u.id === user.id) {
                return { ...u, wxOpenid: openid, wxUnionid: user.wxUnionid, wxSessionKey: user.wxSessionKey, refreshToken: tokens.refreshToken, refreshTokenExpiresAt: tokens.refreshTokenExpiresAt };
            }
            return u;
        });
        await writeUsersDB(updatedUsers);

        res.json({ success: true, user: sanitizeUser(user), isNewUser, ...tokens });
    } catch (err) {
        console.error('[wx-login] failed:', err);
        res.status(500).json({ success: false, message: err?.message || '微信登录失败' });
    }
});

// WeChat Phone Number Binding (phone code → phone number → update user)
app.post('/api/auth/wx-phone', authRequired, async (req, res) => {
    try {
        const { code } = req.body || {};
        if (!code) {
            return res.status(400).json({ success: false, message: '缺少手机号授权码' });
        }
        if (!WX_APPID || !WX_SECRET) {
            return res.status(500).json({ success: false, message: '微信小程序配置缺失' });
        }

        const accessToken = await getWxAccessToken();
        const { data: phoneRes } = await axios.post(
            `https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=${accessToken}`,
            { code }
        );

        if (phoneRes.errcode) {
            console.error('[wx-phone] getPhoneNumber failed:', phoneRes);
            return res.status(400).json({ success: false, message: `获取手机号失败: ${phoneRes.errmsg}` });
        }

        const phoneNumber = phoneRes.phone_info?.purePhoneNumber || phoneRes.phone_info?.phoneNumber || '';
        if (!phoneNumber) {
            return res.status(400).json({ success: false, message: '未获取到手机号' });
        }

        const users = await readUsersDB();
        const existingUser = users.find(u => u.phone === phoneNumber && u.id !== req.user.id);
        if (existingUser) {
            return res.status(400).json({ success: false, message: '该手机号已被其他账号绑定' });
        }

        const updatedUsers = users.map(u => {
            if (u.id === req.user.id) {
                return { ...u, phone: phoneNumber, name: u.name.startsWith('微信用户') ? `用户${phoneNumber.slice(-4)}` : u.name };
            }
            return u;
        });
        await writeUsersDB(updatedUsers);

        const updatedUser = updatedUsers.find(u => u.id === req.user.id);
        res.json({ success: true, user: sanitizeUser(updatedUser), phone: phoneNumber });
    } catch (err) {
        console.error('[wx-phone] failed:', err);
        res.status(500).json({ success: false, message: err?.message || '获取手机号失败' });
    }
});

// WeChat H5 OAuth - 公众号一键登录入口（直接跳转微信授权页）
// 用法：公众号菜单/文章链接设为 https://gzhtest.zndkfx.com/api/auth/wechat-entry?redirect=/home
app.get('/api/auth/wechat-entry', (req, res) => {
    const { redirect } = req.query;

    if (!WX_MP_APPID) {
        return res.status(500).send('微信公众号未配置');
    }

    const callbackUrl = `${config.server.baseUrl}/api/auth/wechat-oauth/callback`;
    const state = redirect || '/home';
    const authUrl = `https://open.weixin.qq.com/connect/oauth2/authorize?appid=${WX_MP_APPID}&redirect_uri=${encodeURIComponent(callbackUrl)}&response_type=code&scope=snsapi_userinfo&state=${encodeURIComponent(state)}#wechat_redirect`;

    res.redirect(authUrl);
});

// WeChat H5 OAuth - 获取公众号授权URL（前端AJAX调用）
app.get('/api/auth/wechat-oauth-url', (req, res) => {
    const { redirectUrl } = req.query;

    if (!WX_MP_APPID || !WX_MP_SECRET) {
        return res.status(500).json({
            success: false,
            message: '微信公众号未配置'
        });
    }

    const redirectUri = encodeURIComponent(`${config.server.baseUrl}/api/auth/wechat-oauth/callback`);
    const state = redirectUrl || 'home';
    const authUrl = `https://open.weixin.qq.com/connect/oauth2/authorize?appid=${WX_MP_APPID}&redirect_uri=${redirectUri}&response_type=code&scope=snsapi_userinfo&state=${encodeURIComponent(state)}#wechat_redirect`;

    res.json({ success: true, authUrl });
});

// WeChat H5 OAuth - 授权回调
app.get('/api/auth/wechat-oauth/callback', async (req, res) => {
    const { code, state } = req.query;

    if (!code) {
        return res.redirect(`${config.server.frontendUrl}/login?error=wechat_auth_failed`);
    }

    try {
        // 1. 通过code获取access_token和openid
        const tokenUrl = `https://api.weixin.qq.com/sns/oauth2/access_token?appid=${WX_MP_APPID}&secret=${WX_MP_SECRET}&code=${code}&grant_type=authorization_code`;
        const { data: tokenData } = await axios.get(tokenUrl);

        if (tokenData.errcode) {
            console.error('[wechat-oauth] token failed:', tokenData);
            return res.redirect(`${config.server.frontendUrl}/login?error=wechat_auth_failed`);
        }

        // 2. 通过access_token和openid获取用户信息
        const userInfoUrl = `https://api.weixin.qq.com/sns/userinfo?access_token=${tokenData.access_token}&openid=${tokenData.openid}&lang=zh_CN`;
        const { data: userInfo } = await axios.get(userInfoUrl);

        if (userInfo.errcode) {
            console.error('[wechat-oauth] userinfo failed:', userInfo);
            return res.redirect(`${config.server.frontendUrl}/login?error=wechat_auth_failed`);
        }

        // 3. 查找或创建用户
        const users = await readUsersDB();
        let user = users.find(u => u.wxOpenid === userInfo.openid);

        if (!user && userInfo.unionid) {
            user = users.find(u => u.wxUnionid === userInfo.unionid);
        }

        if (!user) {
            user = {
                id: randomUUID(),
                phone: '',
                username: `wx_mp_${userInfo.openid.substring(0, 16)}`,
                password: '',
                passwordHash: '',
                name: userInfo.nickname || '微信用户',
                role: 'user',
                avatar: userInfo.headimgurl || '',
                wxOpenid: userInfo.openid,
                wxUnionid: userInfo.unionid || '',
                createTime: new Date().toISOString()
            };
            users.push(user);
            await writeUsersDB(users);
            console.log('[wechat-oauth] new user registered:', user.id);
        } else {
            // 更新头像和昵称（如果有变化）
            if (userInfo.headimgurl && user.avatar !== userInfo.headimgurl) {
                user.avatar = userInfo.headimgurl;
            }
            if (userInfo.nickname && user.name.startsWith('微信用户')) {
                user.name = userInfo.nickname;
            }
            if (!user.wxOpenid) user.wxOpenid = userInfo.openid;
            if (userInfo.unionid && !user.wxUnionid) user.wxUnionid = userInfo.unionid;
        }

        // 4. 生成token
        const tokens = generateTokens(user);
        const updatedUsers = users.map(u => {
            if (u.id === user.id) {
                return { ...u, ...user, refreshToken: tokens.refreshToken, refreshTokenExpiresAt: tokens.refreshTokenExpiresAt };
            }
            return u;
        });
        await writeUsersDB(updatedUsers);

        // 5. 重定向到前端页面，携带token和用户信息
        const decodedState = decodeURIComponent(state || '');
        const redirectPath = decodedState && decodedState !== 'home' ? decodedState : '/home';
        const userData = Buffer.from(JSON.stringify(sanitizeUser(user))).toString('base64');
        const tokenDataEncoded = Buffer.from(JSON.stringify(tokens)).toString('base64');

        // 判断前端URL是否已包含完整路径
        let finalRedirectUrl;
        if (redirectPath.startsWith('http')) {
            const url = new URL(redirectPath);
            url.searchParams.set('wechat_auth', '1');
            url.searchParams.set('user', userData);
            url.searchParams.set('tokens', tokenDataEncoded);
            finalRedirectUrl = url.toString();
        } else {
            finalRedirectUrl = `${config.server.frontendUrl}${redirectPath.startsWith('/') ? '' : '/'}${redirectPath}?wechat_auth=1&user=${userData}&tokens=${tokenDataEncoded}`;
        }

        res.redirect(finalRedirectUrl);
    } catch (err) {
        console.error('[wechat-oauth] callback error:', err.message);
        res.redirect(`${config.server.frontendUrl}/login?error=wechat_auth_error`);
    }
});

// Update User Profile
app.post('/api/user/update', authRequired, async (req, res) => {
    const { id, name, avatar, phone } = req.body || {};
    if (req.user.role !== 'admin' && req.user.id !== id) {
        return res.status(403).json({ success: false, message: '无权限修改该用户' });
    }
    let users = await readUsersDB();
    const index = users.findIndex(u => u.id === id);

    if (index !== -1) {
        users[index] = { ...users[index], name, avatar, phone }; // Keep role and password intact for now
        if (await writeUsersDB(users)) {
            const { password, ...userInfo } = users[index];
            res.json({ success: true, user: userInfo });
        } else {
            res.status(500).json({ success: false, message: 'Failed to update user' });
        }
    } else {
        res.status(404).json({ success: false, message: 'User not found' });
    }
});

// Get Users (Admin only)
app.get('/api/users', authRequired, roleRequired(['admin', 'dsl_admin']), async (req, res) => {
    const users = await readUsersDB();
    const safeUsers = users.map(u => sanitizeUser(u));
    res.json(safeUsers);
});

// Update User Role
app.post('/api/user/role', authRequired, roleRequired(['admin']), async (req, res) => {
    const { id, role } = req.body || {};
    let users = await readUsersDB();
    const index = users.findIndex(u => u.id === id);

    if (index !== -1) {
        const requester = req.user;
        const target = users[index];
        const allowedRoles = ['user', 'admin', 'dsl_admin', 'study_admin'];

        if (requester.phone !== SUPER_ADMIN_PHONE) {
            return res.status(403).json({ success: false, message: '仅超级管理员可调整权限' });
        }
        if (target.phone === SUPER_ADMIN_PHONE) {
            return res.status(400).json({ success: false, message: '超级管理员权限不可修改' });
        }
        if (!allowedRoles.includes(role)) {
            return res.status(400).json({ success: false, message: '非法角色' });
        }

        users[index].role = role;
        if (await writeUsersDB(users)) {
            res.json({ success: true });
        } else {
            res.status(500).json({ success: false, message: 'Failed to update user role' });
        }
    } else {
        res.status(404).json({ success: false, message: 'User not found' });
    }
});

// Study Showcase (merged into services config for ID 9)
app.get('/api/study/showcase', async (req, res) => {
    const config = await readServicesConfig();
    const data = config['9']?.studyShowcase || [];
    res.json({ success: true, data });
});

app.post('/api/study/showcase', authRequired, roleRequired(['admin', 'dsl_admin', 'study_admin']), async (req, res) => {
    const { items } = req.body || {};
    if (!Array.isArray(items)) {
        return res.status(400).json({ success: false, message: 'items must be an array' });
    }

    const config = await readServicesConfig();
    if (!config['9']) config['9'] = {};
    
    config['9'].studyShowcase = items
        .filter(it => it && typeof it === 'object')
        .map(it => ({
            title: String(it.title || '').trim(),
            desc: String(it.desc || '').trim(),
            image: String(it.image || '').trim()
        }))
        .filter(it => it.title || it.desc || it.image);

    if (await writeServicesConfig(config)) {
        res.json({ success: true });
    } else {
        res.status(500).json({ success: false, message: 'Failed to save showcase' });
    }
});

// Services Config (All text editable)
app.get('/api/services/config', async (req, res) => {
    const config = await readServicesConfig();
    res.json({ success: true, data: config });
});

app.post('/api/services/config', authRequired, roleRequired(['admin', 'dsl_admin', 'study_admin']), async (req, res) => {
    const { config } = req.body;
    if (!config || typeof config !== 'object') {
        return res.status(400).json({ success: false, message: 'Invalid config' });
    }

    const role = req.user?.role;
    if (role === 'study_admin') {
        const existing = await readServicesConfig();
        const allowed = config['9'];
        if (!allowed) {
            return res.status(403).json({ success: false, message: '研学管理员仅可修改研学(ID:9)配置' });
        }
        existing['9'] = { ...existing['9'], ...allowed };
        if (await writeServicesConfig(existing)) {
            return res.json({ success: true });
        }
        return res.status(500).json({ success: false, message: 'Failed to save services config' });
    }

    if (await writeServicesConfig(config)) {
        res.json({ success: true });
    } else {
        res.status(500).json({ success: false, message: 'Failed to save services config' });
    }
});

// ─── Dashboard Stats API ─────────────────────────────────────────────
app.get('/api/admin/stats', authRequired, roleRequired(['admin', 'dsl_admin', 'study_admin']), async (req, res) => {
    try {
        const role = req.user?.role;
        const [allApps, allUsers, allCases] = await Promise.all([
            readDB(),
            readUsersDB(),
            readCasesDB()
        ]);

        // Applications visible to this role
        let apps = allApps;
        if (role === 'dsl_admin') {
            apps = apps.filter(item => item.serviceId === '13');
        } else if (role === 'study_admin') {
            apps = apps.filter(item => item.serviceId === '9');
        }

        // Overview counts
        const overview = {
            totalOrders: apps.length,
            pendingOrders: apps.filter(a => !a.status || a.status === '待处理').length,
            processingOrders: apps.filter(a => a.status === '处理中').length,
            completedOrders: apps.filter(a => a.status === '已完成').length,
            totalUsers: allUsers.length,
            totalCases: allCases.length,
            totalCompetition: allApps.filter(a => a.serviceId === '13').length
        };

        // Order trend – last 14 days
        const now = new Date();
        const orderTrend = [];
        for (let i = 13; i >= 0; i--) {
            const d = new Date(now);
            d.setDate(d.getDate() - i);
            const dateStr = d.toISOString().slice(0, 10);
            const count = apps.filter(a => {
                if (!a.createTime) return false;
                return a.createTime.slice(0, 10) === dateStr;
            }).length;
            orderTrend.push({ date: dateStr, count });
        }

        // Competition by role
        const compApps = allApps.filter(a => a.serviceId === '13');
        const competitionByRole = {
            athlete: compApps.filter(a => a.competitionRole === 'athlete').length,
            coach: compApps.filter(a => a.competitionRole === 'coach').length,
            referee: compApps.filter(a => a.competitionRole === 'referee').length,
            club: compApps.filter(a => a.competitionRole === 'club').length
        };

        // User growth – group by month
        const monthMap = {};
        allUsers.forEach(u => {
            if (!u.createTime) return;
            const m = u.createTime.slice(0, 7); // "YYYY-MM"
            monthMap[m] = (monthMap[m] || 0) + 1;
        });
        const userGrowth = Object.entries(monthMap)
            .sort(([a], [b]) => a.localeCompare(b))
            .slice(-12)
            .map(([month, count]) => ({ month, count }));

        // Status distribution
        const statusDist = {};
        apps.forEach(a => {
            const s = a.status || '待处理';
            statusDist[s] = (statusDist[s] || 0) + 1;
        });

        res.json({
            success: true,
            data: { overview, orderTrend, competitionByRole, userGrowth, statusDist }
        });
    } catch (err) {
        console.error('[admin/stats] failed:', err);
        res.status(500).json({ success: false, message: '获取统计数据失败' });
    }
});

// Submit Application
app.post('/api/submit', authOptional, async (req, res) => {
    const application = req.body;
    if (req.user && !application.userId) {
        application.userId = req.user.id;
    }
    application.id = Date.now().toString(); // Simple ID
    application.createTime = new Date().toISOString(); // Server timestamp

    const db = await readDB();
    db.push(application);
    await writeDB(db);

    res.json({ success: true, message: 'Application submitted successfully', id: application.id });
});

// Update Application Status
app.post('/api/update', async (req, res) => {
    const { id, status } = req.body;
    let db = await readDB();
    const index = db.findIndex(item => item.id == id);

    if (index !== -1) {
        db[index].status = status;
        if (await writeDB(db)) {
            res.json({ success: true });
        } else {
            res.status(500).json({ success: false, message: 'Failed to write to database' });
        }
    } else {
        res.status(404).json({ success: false, message: 'Application not found' });
    }
});

// Get Applications (with optional filtering)
app.get('/api/list', authRequired, async (req, res) => {
    const { startDate, endDate } = req.query;
    let db = await readDB();

    const role = req.user?.role;
    const userId = req.user?.id;

    // DSL管理员权限：仅支持查阅管理无人机赛事(ID 13)信息
    if (role === 'dsl_admin') {
        db = db.filter(item => item.serviceId === '13');
    } else if (role === 'study_admin') {
        db = db.filter(item => item.serviceId === '9');
    } else if (userId && role !== 'admin') {
        // Filter by User ID if provided (for regular users)
        db = db.filter(item => item.userId === userId);
    }

    if (startDate || endDate) {
        db = db.filter(item => {
            const itemDate = new Date(item.createTime).getTime();
            const start = startDate ? new Date(startDate).getTime() : 0;
            const end = endDate ? new Date(endDate).getTime() : Infinity;
            return itemDate >= start && itemDate <= end;
        });
    }

    // Sort by newest first
    db.sort((a, b) => new Date(b.createTime) - new Date(a.createTime));

    res.json(db);
});

// Export to Excel
app.get('/api/export', authRequired, roleRequired(['admin', 'dsl_admin', 'study_admin']), async (req, res) => {
    const { startDate, endDate, ids } = req.query;
    let db = await readDB();
    const role = req.user?.role;

    // DSL管理员权限过滤
    if (role === 'dsl_admin') {
        db = db.filter(item => item.serviceId === '13');
    } else if (role === 'study_admin') {
        db = db.filter(item => item.serviceId === '9');
    }

    // 如果指定了IDS，则只导出选中的项
    if (ids) {
        const idList = ids.split(',');
        db = db.filter(item => idList.includes(item.id.toString()));
    } else if (startDate || endDate) {
        db = db.filter(item => {
            const itemDate = new Date(item.createTime).getTime();
            const start = startDate ? new Date(startDate).getTime() : 0;
            const end = endDate ? new Date(endDate).getTime() : Infinity;
            return itemDate >= start && itemDate <= end;
        });
    }
    
    // Flatten data for Excel
    const excelData = db.map(item => {
        const baseData = {
            ID: item.id,
            提交时间: new Date(item.createTime).toLocaleString(),
            服务名称: item.serviceName || '',
            状态: item.status || '待处理'
        };

        if (item.serviceId === '13') {
            return {
                ...baseData,
                注册号: item.regNo || '',
                注册角色: item.competitionRoleText || '',
                单位名称: item.companyName || '',
                姓名: item.name || '',
                性别: item.gender === 'male' ? '男' : '女',
                证件号: item.idCard || '',
                组别: item.competitionGroup || item.athleteGroup || '',
                参赛项目: item.competitionProject || '',
                联系电话: item.phone || item.managerPhone || item.contactPhone || '',
                电子邮箱: item.email || '',
                所在地: item.location || '',
                等级: item.level || '',
                有效期: item.validDate || '',
                负责人: item.manager || '',
                主要对接人: item.contactPerson || '',
                备注: item.remark || ''
            };
        }

        return {
            ...baseData,
            联系人: item.contactName || '',
            联系电话: item.contactPhone || '',
            学员姓名: item.traineeName || '',
            学员电话: item.traineePhone || '',
            性别: item.traineeGender === 'male' ? '男' : (item.traineeGender === 'female' ? '女' : ''),
            身份证号: item.traineeIdCard || '',
            考试机型: item.examModel || '',
            证照级别: item.licenseLevel || '',
            客户类型: item.customerType === 'enterprise' ? '企业' : '个人',
            企业名称: item.companyName || '',
            货物类型: item.cargoType || '',
            起运地: item.startAddress || '',
            目的地: item.endAddress || ''
        };
    });

    const wb = XLSX.utils.book_new();
    const ws = XLSX.utils.json_to_sheet(excelData);
    XLSX.utils.book_append_sheet(wb, ws, "Applications");

    const buffer = XLSX.write(wb, { type: 'buffer', bookType: 'xlsx' });

    res.setHeader('Content-Disposition', `attachment; filename="export_${Date.now()}.xlsx"`);
    res.setHeader('Content-Type', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
    res.send(buffer);
});

// ========== 服务评价 API ==========

const REVIEW_SECTIONS = ['yanxue', 'sale', 'park'];

// 获取板块对应的可选课程/服务列表（供评价时选择）
app.get('/api/reviews/courses', async (req, res) => {
    try {
        const { section } = req.query;
        const config = await readServicesConfig();
        let courses = [];

        if (section === 'yanxue') {
            const pkgs = config?.['9']?.packages || {};
            courses = Object.keys(pkgs).map(key => ({
                id: key,
                name: pkgs[key].name || key
            })).filter(c => c.name);
        }

        res.json({ success: true, data: courses });
    } catch (err) {
        res.json({ success: true, data: [] });
    }
});

// 获取已审核通过的评价（公开）
app.get('/api/reviews', async (req, res) => {
    try {
        const reviews = await readReviewsDB();
        const { section } = req.query;

        let approved = reviews.filter(r => r.status === 'approved');

        if (section && REVIEW_SECTIONS.includes(section)) {
            approved = approved.filter(r => r.section === section);
        }

        approved.sort((a, b) => new Date(b.createTime) - new Date(a.createTime));

        res.json({ success: true, data: approved });
    } catch (err) {
        res.status(500).json({ success: false, message: '获取评价失败' });
    }
});

// 提交评价（需登录）
app.post('/api/reviews', authRequired, async (req, res) => {
    try {
        const { section, rating, content, isAnonymous, courseName, images } = req.body;

        if (!section || !REVIEW_SECTIONS.includes(section)) {
            return res.status(400).json({ success: false, message: '无效的评价板块' });
        }
        if (!rating || rating < 1 || rating > 5 || !Number.isInteger(rating)) {
            return res.status(400).json({ success: false, message: '评分必须为1-5的整数' });
        }
        if (!content || typeof content !== 'string' || content.trim().length === 0) {
            return res.status(400).json({ success: false, message: '评价内容不能为空' });
        }
        if (content.length > 500) {
            return res.status(400).json({ success: false, message: '评价内容不能超过500字' });
        }
        // 校验图片：最多9张，每项必须为字符串URL
        let validImages = [];
        if (Array.isArray(images)) {
            validImages = images.filter(u => typeof u === 'string' && u.trim()).slice(0, 9);
        }

        const reviews = await readReviewsDB();

        // 根据是否匿名决定显示的名称和头像
        const displayName = isAnonymous ? '匿名用户' : (req.user.name || '匿名用户');
        const displayAvatar = isAnonymous ? '' : (req.user.avatar || '');

        const newReview = {
            id: randomUUID(),
            userId: req.user.id,
            userName: displayName,
            userAvatar: displayAvatar,
            isAnonymous: isAnonymous || false,
            section,
            rating,
            content: content.trim(),
            courseName: (typeof courseName === 'string' && courseName.trim()) ? courseName.trim() : '',
            images: validImages,
            status: 'pending',
            createTime: new Date().toISOString()
        };

        reviews.push(newReview);
        await writeReviewsDB(reviews);

        res.json({ success: true, message: '评价提交成功，等待审核' });
    } catch (err) {
        res.status(500).json({ success: false, message: '提交评价失败' });
    }
});

// Get Cases (with pagination)
app.get('/api/cases', async (req, res) => {
    let cases = await readCasesDB();
    const { categoryId, page, limit } = req.query;

    // Filter by category
    if (categoryId && categoryId !== '0') {
        cases = cases.filter(c => c.categoryId == categoryId);
    }

    // Sort newest first
    cases.sort((a, b) => new Date(b.date) - new Date(a.date)); // Assuming date field or createTime. Existing code didn't sort explicitly but it's good practice. Or use 'id' which is timestamp.
    // Actually existing code didn't sort cases.json. I'll sort by id (timestamp) descending.
    cases.sort((a, b) => b.id - a.id);

    if (page && limit) {
        const p = parseInt(page);
        const l = parseInt(limit);
        const start = (p - 1) * l;
        const end = start + l;
        const paginated = cases.slice(start, end);
        res.json({
            data: paginated,
            total: cases.length,
            page: p,
            limit: l
        });
    } else {
        res.json(cases);
    }
});

// Create Case
app.post('/api/cases/create', async (req, res) => {
    const newCase = req.body;
    newCase.id = Date.now(); // Simple ID generation
    let cases = await readCasesDB();
    cases.unshift(newCase); // Add to top
    if (await writeCasesDB(cases)) {
        res.json({ success: true, id: newCase.id });
    } else {
        res.status(500).json({ success: false, message: 'Failed to create case' });
    }
});

// Update Case
app.post('/api/cases/update', async (req, res) => {
    const updatedCase = req.body;
    let cases = await readCasesDB();
    const index = cases.findIndex(c => c.id == updatedCase.id);

    if (index !== -1) {
        cases[index] = { ...cases[index], ...updatedCase };
        if (await writeCasesDB(cases)) {
            res.json({ success: true });
        } else {
            res.status(500).json({ success: false, message: 'Failed to write to database' });
        }
    } else {
        // Optional: Add new case if not found, or return 404. For now, let's assume editing existing.
        // But user might want to add cases later. Let's support simple update for now.
        res.status(404).json({ success: false, message: 'Case not found' });
    }
});

// Delete Case
app.post('/api/cases/delete', async (req, res) => {
    const { id } = req.body;
    let cases = await readCasesDB();
    const index = cases.findIndex(c => c.id == id);

    if (index !== -1) {
        cases.splice(index, 1);
        if (await writeCasesDB(cases)) {
            res.json({ success: true });
        } else {
            res.status(500).json({ success: false, message: 'Failed to write to database' });
        }
    } else {
        res.status(404).json({ success: false, message: 'Case not found' });
    }
});

// ==================== 案例分类 CRUD ====================
// 获取分类列表（公开）
app.get('/api/case-categories', async (req, res) => {
    const categories = await readCaseCategoriesDB();
    res.json(Array.isArray(categories) ? categories : []);
});

// 新增分类
app.post('/api/case-categories/create', async (req, res) => {
    const { name, service } = req.body || {};
    if (!name || !String(name).trim()) {
        return res.status(400).json({ success: false, message: '分类名称不能为空' });
    }
    const categories = (await readCaseCategoriesDB()) || [];
    // 自动分配 id：取现有最大 id + 1，且不小于 1
    const maxId = categories.reduce((m, c) => Math.max(m, Number(c.id) || 0), 0);
    const newCat = {
        id: maxId + 1,
        name: String(name).trim(),
        service: service ? String(service).trim() : String(name).trim()
    };
    categories.push(newCat);
    if (await writeCaseCategoriesDB(categories)) {
        res.json({ success: true, data: newCat });
    } else {
        res.status(500).json({ success: false, message: '保存失败' });
    }
});

// 更新分类
app.post('/api/case-categories/update', async (req, res) => {
    const { id, name, service } = req.body || {};
    if (id == null) {
        return res.status(400).json({ success: false, message: '缺少分类 id' });
    }
    const categories = (await readCaseCategoriesDB()) || [];
    const index = categories.findIndex(c => Number(c.id) === Number(id));
    if (index === -1) {
        return res.status(404).json({ success: false, message: '分类不存在' });
    }
    if (name !== undefined) categories[index].name = String(name).trim();
    if (service !== undefined) categories[index].service = String(service).trim();
    if (await writeCaseCategoriesDB(categories)) {
        res.json({ success: true, data: categories[index] });
    } else {
        res.status(500).json({ success: false, message: '保存失败' });
    }
});

// 删除分类（若有案例引用该分类则拒绝）
app.post('/api/case-categories/delete', async (req, res) => {
    const { id } = req.body || {};
    if (id == null) {
        return res.status(400).json({ success: false, message: '缺少分类 id' });
    }
    const cases = (await readCasesDB()) || [];
    const inUse = cases.some(c => Number(c.categoryId) === Number(id));
    if (inUse) {
        return res.status(400).json({ success: false, message: '该分类下仍有案例，无法删除' });
    }
    const categories = (await readCaseCategoriesDB()) || [];
    const index = categories.findIndex(c => Number(c.id) === Number(id));
    if (index === -1) {
        return res.status(404).json({ success: false, message: '分类不存在' });
    }
    categories.splice(index, 1);
    if (await writeCaseCategoriesDB(categories)) {
        res.json({ success: true });
    } else {
        res.status(500).json({ success: false, message: '保存失败' });
    }
});

// 挂载管理后台路由 (必须在SPA路由处理器之前)
app.use('/api/admin', adminRouter);

// 挂载医疗配送模块路由
app.use('/api/medical', medicalRouter);

// Handle SPA routing: Serve index.html for all non-API routes
app.get('*', (req, res) => {
    // If it's an API request that wasn't handled above, return 404
    if (req.path.startsWith('/api/')) {
        return res.status(404).json({ success: false, message: 'API not found' });
    }
    
    // If it looks like a static file (has extension) but wasn't found in public/, return 404
    // This prevents returning index.html for missing js/css files
    if (req.path.includes('.')) {
        return res.status(404).send('Not Found');
    }

    // Otherwise serve index.html for client-side routing (Vue Router)
    res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

// 404 handler for API routes
app.use(notFoundHandler);

// Global error handler
app.use(errorHandler);

async function bootstrap() {
    // 打印配置信息
    printConfig();

    // 初始化存储
    await initStorage();
    await ensureDefaultCases();
    await ensureAdminUsers();

    // 启动医疗配送超时检查任务
    startTimeoutChecker();

    // 启动服务器
    app.listen(PORT, () => {
        logger.info(`Server is running on http://localhost:${PORT}`);
        logger.info(`Environment: ${config.server.env}`);
        logger.info(`Database: ${config.database.usePostgres ? 'PostgreSQL' : 'JSON File'}`);
    });
}

bootstrap().catch(err => {
    logger.error('Failed to start server', { error: err.message, stack: err.stack });
    process.exit(1);
});
