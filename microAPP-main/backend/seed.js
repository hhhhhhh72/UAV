const fs = require('fs');
const path = require('path');

const DB_FILE = path.join(__dirname, 'data.json');

const testData = [
    {
        "id": "1732450000001",
        "createTime": "2025-11-24T10:30:00.000Z",
        "contactName": "张三",
        "contactPhone": "13800138000",
        "serviceName": "无人机物流服务",
        "serviceId": "1",
        "customerType": "personal",
        "cargoType": "生鲜食品",
        "cargoWeight": "5",
        "startAddress": "杭州市余杭区",
        "endAddress": "杭州市西湖区",
        "status": "待处理"
    },
    {
        "id": "1732450000002",
        "createTime": "2025-11-24T11:15:00.000Z",
        "contactName": "李四企业",
        "contactPhone": "13900139000",
        "serviceName": "政务巡检服务",
        "serviceId": "2",
        "inspectionType": "河道巡检",
        "inspectionArea": "运河段",
        "status": "已完成"
    }
];

fs.writeFileSync(DB_FILE, JSON.stringify(testData, null, 2));
console.log('Test data seeded successfully.');
