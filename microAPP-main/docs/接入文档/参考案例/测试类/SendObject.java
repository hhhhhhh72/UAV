/**
 * Copyright (C), 2021, JieYi Software System Co., Ltd.
 * All rights reserved.
 * FileName:       SendObject
 *
 * @Description:
 * @author: victor
 * @version: V1.0
 * Create Date:    2021/10/12 17:03
 * <p>
 * Modification History:
 * Date         Author        Version        Discription
 * -------------------------------------------------------
 * 2021/10/12      victor          1.0             add
 */
package com.jieyi.a;

import java.util.Map;

/**
 * @author victor
 * @Description:发送对象
 * @date 2021年10月12日 17:03
 */
public class SendObject {
    /**
     * 接入机构
     */
    private String joininstid;
    /**
     * 报文唯一流水号
     */
    private String joininstssn;
    /**
     * 报文发送日期
     */
    private String reqdate;
    /**
     * 报文发送时间
     */
    private String reqtime;
    /**
     * 请求头对象
     */
    private Map<String, Object> hDataMap;
    /**
     * 请求体对象
     */
    private Map<String, Object> dataMap;
    /**
     * 报文加密密钥
     */
    private String dgKey;
    /**
     * 签名私钥
     */
    private String signPrivateKey;
    /**
     * 调用URL
     */
    private String sendUrl;

    public Map<String, Object> gethDataMap() {
        return hDataMap;
    }

    public void sethDataMap(Map<String, Object> hDataMap) {
        this.hDataMap = hDataMap;
    }

    public Map<String, Object> getDataMap() {
        return dataMap;
    }

    public void setDataMap(Map<String, Object> dataMap) {
        this.dataMap = dataMap;
    }

    public String getDgKey() {
        return dgKey;
    }

    public void setDgKey(String dgKey) {
        this.dgKey = dgKey;
    }

    public String getSignPrivateKey() {
        return signPrivateKey;
    }

    public void setSignPrivateKey(String signPrivateKey) {
        this.signPrivateKey = signPrivateKey;
    }

    public String getJoininstid() {
        return joininstid;
    }

    public void setJoininstid(String joininstid) {
        this.joininstid = joininstid;
    }

    public String getJoininstssn() {
        return joininstssn;
    }

    public void setJoininstssn(String joininstssn) {
        this.joininstssn = joininstssn;
    }

    public String getReqdate() {
        return reqdate;
    }

    public void setReqdate(String reqdate) {
        this.reqdate = reqdate;
    }

    public String getReqtime() {
        return reqtime;
    }

    public void setReqtime(String reqtime) {
        this.reqtime = reqtime;
    }

    public String getSendUrl() {
        return sendUrl;
    }

    public void setSendUrl(String sendUrl) {
        this.sendUrl = sendUrl;
    }
}
