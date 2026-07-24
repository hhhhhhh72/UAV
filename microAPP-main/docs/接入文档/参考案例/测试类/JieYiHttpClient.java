/**
 * Copyright (C), 2021, JieYi Software System Co., Ltd.
 * All rights reserved.
 * FileName:       JieYiHttpClient
 *
 * @Description:
 * @author: victor
 * @version: V1.0
 * Create Date:    2021/10/12 17:02
 * <p>
 * Modification History:
 * Date         Author        Version        Discription
 * -------------------------------------------------------
 * 2021/10/12      victor          1.0             add
 */
package com.jieyi.a;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.jieyi.sdk.RecvObject;
import com.jieyi.sdk.SendObject;
import com.jieyi.util.HttpClientUtil;
import jieyi.tools.algorithmic.SHA1Util;
import jieyi.tools.algorithmic.sm.SM2Utils;
import jieyi.tools.algorithmic.sm.SM4Utils;
import jieyi.tools.algorithmic.sm.Util;
import jieyi.tools.util.NSUtil;

import javax.xml.bind.DatatypeConverter;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * @author victor
 * @Description:
 * @date 2021年10月12日 17:02
 */
public class JieYiHttpClient {

    public RecvObject sendData(SendObject sendObject) throws Exception {
        Map<String, Object> hDataMap = sendObject.gethDataMap();
        Map<String, Object> dataMap = sendObject.getDataMap();

        Gson gson = new GsonBuilder().disableHtmlEscaping().create();

        Map map = new HashMap<>(16);
        map.put("joininstid", sendObject.getJoininstid());
        map.put("joininstssn", sendObject.getJoininstssn());
        map.put("reqdate", sendObject.getReqdate());
        map.put("reqtime", sendObject.getReqtime());
        map.put("hdata", sendObject.gethDataMap());
        map.put("data", sendObject.getDataMap());

        System.out.println("enc bef dg:" + gson.toJson(map));
        String sm4_secretKey = sendObject.getDgKey();
        System.out.println("dg key:" + sm4_secretKey);
        SM4Utils sm4 = new SM4Utils();
        sm4.secretKey = sm4_secretKey;
        sm4.hexString = true;

        //加密hdata和data
        String hdataStr = gson.toJson(sendObject.gethDataMap());
        String dataStr = gson.toJson(sendObject.getDataMap());
        byte[] hdataEncByte = sm4.encryptData_ECB(hdataStr.getBytes("utf-8"));
        String hdataEnc = DatatypeConverter.printBase64Binary(hdataEncByte);
        byte[] dataEncByte = sm4.encryptData_ECB(dataStr.getBytes("utf-8"));
        String dataEnc = DatatypeConverter.printBase64Binary(dataEncByte);

        map.put("hdataenc", hdataEnc);
        map.put("dataenc", dataEnc);
        map.remove("hdata");
        map.remove("data");
        //签名
        List<Map.Entry<String, String>> mappingList = null;
        // 通过ArrayList构造函数把map.entrySet()转换成list
        mappingList = new ArrayList<Map.Entry<String, String>>(map.entrySet());
        Collections.sort(mappingList, new Comparator<Map.Entry<String, String>>() {
            @Override
            public int compare(Map.Entry<String, String> mapping1, Map.Entry<String, String> mapping2) {
                return mapping1.getKey().compareTo(mapping2.getKey());
            }
        });

        String strForSign = "";
        for (Map.Entry<String, String> m : mappingList) {
            strForSign += m.getKey() + m.getValue();
        }
        System.out.println("strForSign:" + strForSign);
        String strForSignSha1 = SHA1Util.calc(strForSign);
        System.out.println("strForSignSha1:" + strForSignSha1);

        String sendPrivatekey = sendObject.getSignPrivateKey();
        System.out.println("sign private key:" + sendPrivatekey);
        String joininstidStr = (String) map.get("joininstid");
        byte[] signByte = SM2Utils.sign(joininstidStr.getBytes("utf-8"),
                Util.hexToByte(sendPrivatekey),
                Util.hexToByte(strForSignSha1));
        String sign = DatatypeConverter.printBase64Binary(signByte);
        map.put("sign", sign);
        System.out.println("enc aft dg:" + gson.toJson(map));

        long start1 = System.currentTimeMillis();

        //发送
        String strRet = HttpClientUtil.httpPostJson(sendObject.getSendUrl(), gson.toJson(map));
        long end1 = System.currentTimeMillis();
        Map strMap = gson.fromJson(strRet, Map.class);

        System.out.println("enc bef retdg:" + strRet);
        String dataenc = (String) strMap.get("dataenc");
        if (NSUtil.isEmpty(dataenc)) {
            //dataenc 为空，不处理
        } else {
            sm4 = new SM4Utils();
            sm4.secretKey = sendObject.getDgKey();
            sm4.hexString = true;
            byte[] plainByte = sm4.decryptData_ECB(DatatypeConverter.parseBase64Binary(dataenc));
            System.out.println("xxx:" + new String(plainByte, "utf-8").trim());
            strMap.put("data", gson.fromJson(new String(plainByte, "utf-8").trim(), Map.class));
            strMap.remove("dataenc");
            System.out.println("enc aft retdg:" + gson.toJson(strMap));
        }

        System.out.println("cost:" + (end1 - start1) + "ms");
        RecvObject recvObject = new RecvObject();
        recvObject.setResult((String) strMap.get("result"));
        recvObject.setResultdesc((String) strMap.get("resultdesc"));
        recvObject.setData((Map<String, Object>) strMap.get("data"));
        return recvObject;
    }
}
