package com.jieyi.a;

import com.jieyi.sdk.SendObject;
import jieyi.tools.algorithmic.RSAUtil;
import jieyi.tools.algorithmic.SHA1Util;
import jieyi.tools.util.DateUtil;
import jieyi.tools.util.StringUtil;

import java.security.interfaces.RSAPublicKey;
import java.util.HashMap;
import java.util.Map;

public class TestComm {

    public static SendObject procSendData() throws Exception {
        //String joininstid = "01642220";
        String joininstid = "10000001";
        String signPrivateKey = "47fc79d3c7aa472c14a426a296030eb710edbbaa6b4ccf08eec0e9b8a4f71ddf";
        String dgKey = "12345678901234561234567890123456";
        String instid = "00000001";
        String mchntid = "000000010001";
        String chnlid = "04";
        String datatime = DateUtil.getCurrentTime24SSS();

        SendObject sendObject = new SendObject();
        sendObject.setJoininstid(joininstid);
        sendObject.setJoininstssn(StringUtil.getRandomStringAccordingSystemtimeForNumberFlag(20, 0));
        sendObject.setReqdate(datatime.substring(0, 8));
        sendObject.setReqtime(datatime.substring(8, 14));

        Map<String, Object> hDataMap = new HashMap<>(16);
        hDataMap.put("instid", instid);
        hDataMap.put("mchntid", mchntid);
        hDataMap.put("chnlid", chnlid);



        sendObject.sethDataMap(hDataMap);
        sendObject.setDgKey(dgKey);
        sendObject.setSignPrivateKey(signPrivateKey);
        return sendObject;
    }


    public static String otpEnc(String otp, String passWord) throws Exception {
        String calc = SHA1Util.calc(passWord);
        System.out.println("SHA1Util:" + calc);
        byte[] plainText = StringUtil.hexStringToBytes(calc);
        System.out.println("原始数据:" + StringUtil.bytesToHexString(plainText).toUpperCase());
        byte radix = 16;
        RSAPublicKey rsaPublicKey = RSAUtil.loadPublicKey(otp,
                "10001", radix);
        byte[] cipherText = RSAUtil.publicKeyEncrypt("RSA", "ECB", "PKCS1Padding", rsaPublicKey, plainText);
        return StringUtil.bytesToHexString(cipherText).toUpperCase();
    }
}
