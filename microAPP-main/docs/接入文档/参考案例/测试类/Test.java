package com.jieyi.a;


import com.jieyi.sdk.JieYiHttpClient;
import com.jieyi.sdk.SendObject;

import java.util.HashMap;
import java.util.Map;

public class Test extends TestComm {
    public static void main(String[] args) throws Exception {
        //请求地址
        String url = "";
        SendObject sendObject = procSendData();
        sendObject.setSendUrl(url);
        Map<String, Object> dataMap = new HashMap<>(16);
        //请求参数
        dataMap.put("txntype", "2001");


        sendObject.setDataMap(dataMap);
        new JieYiHttpClient().sendData(sendObject);

    }
}
