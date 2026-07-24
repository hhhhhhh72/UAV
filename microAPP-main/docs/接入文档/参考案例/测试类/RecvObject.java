/**
 * Copyright (C), 2021, JieYi Software System Co., Ltd.
 * All rights reserved.
 * FileName:       RecvObject
 *
 * @Description:
 * @author: victor
 * @version: V1.0
 * Create Date:    2021/10/13 12:42
 * <p>
 * Modification History:
 * Date         Author        Version        Discription
 * -------------------------------------------------------
 * 2021/10/13      victor          1.0             add
 */
package com.jieyi.a;

import java.util.Map;

/**
 * @author victor
 * @Description:接收对象
 * @date 2021年10月13日 12:42
 */
public class RecvObject {
    /**
     * 错误码
     */
    private String result;
    /**
     * 错误码描述
     */
    private String resultdesc;
    /**
     * 返回对象
     */
    private Map<String, Object> data;

    public String getResult() {
        return result;
    }

    public void setResult(String result) {
        this.result = result;
    }

    public String getResultdesc() {
        return resultdesc;
    }

    public void setResultdesc(String resultdesc) {
        this.resultdesc = resultdesc;
    }

    public Map<String, Object> getData() {
        return data;
    }

    public void setData(Map<String, Object> data) {
        this.data = data;
    }
}
