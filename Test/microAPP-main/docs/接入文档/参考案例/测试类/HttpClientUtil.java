package com.jieyi.a;

import com.google.gson.Gson;
import com.jieyi.util.SSLClient;
import org.apache.http.HttpEntity;
import org.apache.http.HttpResponse;
import org.apache.http.NameValuePair;
import org.apache.http.client.ClientProtocolException;
import org.apache.http.client.HttpClient;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.conn.ssl.NoopHostnameVerifier;
import org.apache.http.conn.ssl.SSLConnectionSocketFactory;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.entity.mime.HttpMultipartMode;
import org.apache.http.entity.mime.MultipartEntityBuilder;
import org.apache.http.entity.mime.content.FileBody;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.message.BasicNameValuePair;
import org.apache.http.util.EntityUtils;

import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;
import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.UnsupportedEncodingException;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLDecoder;
import java.nio.charset.Charset;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;
import java.util.List;
import java.util.Map;

//import org.apache.http.entity.mime.HttpMultipartMode;
//import org.apache.http.entity.mime.MultipartEntityBuilder;
//import org.apache.http.entity.mime.content.FileBody;

public class HttpClientUtil {


    public static String sendMsgAndRecv(String url, List<NameValuePair> list) throws Exception {
        String sRet = "-1";
        CloseableHttpClient httpClient = getHttpClient();
        try {
            HttpPost post = new HttpPost(url);
            // HttpPost post = new
            // HttpPost("http://127.0.0.1:8080/jytestreport/restful/biz.do"); //
            // 这里用上本机的某个工程做测试
            // 创建参数列表
            // List<NameValuePair> list = new ArrayList<NameValuePair>();
            // list.add(new BasicNameValuePair("j_username", "admin"));
            // list.add(new BasicNameValuePair("j_password", "admin"));
            // url格式编码
            UrlEncodedFormEntity uefEntity = new UrlEncodedFormEntity(list, "UTF-8");
            post.setEntity(uefEntity);
            System.out.println("POST 请求...." + post.getURI());
            // 执行请求
            CloseableHttpResponse httpResponse = httpClient.execute(post);
            try {
                HttpEntity entity = httpResponse.getEntity();
                System.out.println(httpResponse.getStatusLine().getStatusCode());
                if (null != entity) {
                    System.out.println("-------------------------------------------------------");
                    sRet = EntityUtils.toString(entity);
                    System.out.println(sRet);
                    System.out.println("-------------------------------------------------------");

                }
            } finally {
                httpResponse.close();
            }

        } catch (Exception e) {
            e.printStackTrace();
            throw e;
        } finally {
            try {
                closeHttpClient(httpClient);
            } catch (Exception e) {
                e.printStackTrace();
            }
        }

        return sRet;
    }

    private static CloseableHttpClient getHttpClient() {
        // 代理 RequestConfig.custom().setProxy(proxy)
//        HttpHost proxy = new HttpHost("127.0.0.1", 8734);

        RequestConfig defaultRequestConfig = RequestConfig.custom().setSocketTimeout(10000).setConnectTimeout(10000)
                .setSocketTimeout(10000).build();
        CloseableHttpClient httpclient = HttpClients.custom().setDefaultRequestConfig(defaultRequestConfig).build();
        return httpclient;
    }

    private static void closeHttpClient(CloseableHttpClient client) throws IOException {
        if (client != null) {
            client.close();
        }
    }

    /**
     * post请求
     *
     * @param url  url地址
     * @param str1 参数
     * @return
     */
    public static String httpPostJson(String url, String str1) throws IOException {
        // post请求返回结果
        System.out.println("url:" + url);
        //DefaultHttpClient httpClient = new DefaultHttpClient();
        CloseableHttpClient httpClient = getHttpClient();
        if (url.startsWith("https")) {
            httpClient = (CloseableHttpClient) wrapClient(httpClient);
        }
        // JSONObject jsonResult = null;
        HttpPost method = new HttpPost(url);
        String str = null;
        try {
            // 解决中文乱码问题
            StringEntity entity = new StringEntity(str1, "utf-8");
            entity.setContentEncoding("UTF-8");
            entity.setContentType("application/json");
            method.setEntity(entity);

            HttpResponse result = httpClient.execute(method);
            url = URLDecoder.decode(url, "UTF-8");
            /** 请求发送成功，并得到响应 **/
            // System.out.println("EntityUtils.toString(result.getEntity()):" +
            // EntityUtils.toString(result.getEntity()));
            if (result.getStatusLine().getStatusCode() == 200) {
                try {
                    /** 读取服务器返回过来的json字符串数据 **/
                    String s1 = EntityUtils.toString(result.getEntity());
                    //str = new String(s1.getBytes("iso-8859-1"), "utf-8");

                    str = s1;
                } catch (Exception e) {
                    e.printStackTrace();
                }
            } else {
                System.out.println("result.getStatusLine().getStatusCode():" + result.getStatusLine().getStatusCode());
            }
        } catch (IOException e) {
            e.printStackTrace();
            throw e;
        } finally {
            try {
                if (null != method) {
                    method.releaseConnection();
                    method = null;
                }

                closeHttpClient(httpClient);
                httpClient = null;
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
        return str;
    }

    /**
     * post请求,百度智能云
     *
     * @param url  url地址
     * @param str1 参数
     * @return
     */
    public static String httpPostJsonBaidu(String url, String str1) throws IOException {
        // post请求返回结果
        CloseableHttpClient httpClient = getHttpClient();
        HttpPost method = new HttpPost(url);
        String str = null;
        try {
            method.addHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8");

            StringEntity entity = new StringEntity(str1, "utf-8");
            entity.setContentEncoding("UTF-8");
            entity.setContentType("application/x-www-form-urlencoded; charset=UTF-8");
            method.setEntity(entity);


            HttpResponse result = httpClient.execute(method);
            url = URLDecoder.decode(url, "UTF-8");
            if (result.getStatusLine().getStatusCode() == 200) {
                try {
                    /** 读取服务器返回过来的json字符串数据 **/
                    String s1 = EntityUtils.toString(result.getEntity());
                    str = s1;
                } catch (Exception e) {
                    e.printStackTrace();
                }
            } else {
                System.out.println("result.getStatusLine().getStatusCode():" + result.getStatusLine().getStatusCode());
            }
        } catch (IOException e) {
            e.printStackTrace();
            throw e;
        } finally {
            try {
                if (null != method) {
                    method.releaseConnection();
                    method = null;
                }

                closeHttpClient(httpClient);
                httpClient = null;
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
        return str;
    }


    /**
     * post带header请求
     *
     * @param url  url地址
     * @param str1 参数
     * @return
     */
    public static String httpPostJsonWithHeader(String url, String str1, Map<String, String> headerMap) throws IOException {
        // post请求返回结果
        //DefaultHttpClient httpClient = new DefaultHttpClient();
        CloseableHttpClient httpClient = getHttpClient();
        if (url.startsWith("https")) {
            httpClient = (CloseableHttpClient) wrapClient(httpClient);
        }
        // JSONObject jsonResult = null;
        HttpPost method = new HttpPost(url);
        String str = null;
        try {
            // 解决中文乱码问题
            StringEntity entity = new StringEntity(str1, "utf-8");
            entity.setContentEncoding("UTF-8");
            entity.setContentType("application/json");
            method.setEntity(entity);
            Iterator headerIterator = headerMap.entrySet().iterator();          //循环增加header
            while (headerIterator.hasNext()) {
                Map.Entry<String, String> elem = (Map.Entry<String, String>) headerIterator.next();
                method.addHeader(elem.getKey(), elem.getValue());
            }

            HttpResponse result = httpClient.execute(method);
            url = URLDecoder.decode(url, "UTF-8");
            /** 请求发送成功，并得到响应 **/
            // System.out.println("EntityUtils.toString(result.getEntity()):" +
            // EntityUtils.toString(result.getEntity()));
            if (result.getStatusLine().getStatusCode() == 200) {
                try {
                    /** 读取服务器返回过来的json字符串数据 **/
                    String s1 = EntityUtils.toString(result.getEntity());
                    //str = new String(s1.getBytes("iso-8859-1"), "utf-8");

                    str = s1;
                } catch (Exception e) {
                    e.printStackTrace();
                }
            } else {
                System.out.println("result.getStatusLine().getStatusCode():" + result.getStatusLine().getStatusCode());
            }
        } catch (IOException e) {
            e.printStackTrace();
            throw e;
        } finally {
            try {
                if (null != method) {
                    method.releaseConnection();
                    method = null;
                }

                closeHttpClient(httpClient);
                httpClient = null;
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
        return str;
    }

    public static String httpPostJson123(String url, String str1, Map Header) throws IOException {
        // post请求返回结果
        //DefaultHttpClient httpClient = new DefaultHttpClient();
        CloseableHttpClient httpClient = getHttpClient();
        if (url.startsWith("https")) {
            httpClient = (CloseableHttpClient) wrapClient(httpClient);
        }
        // JSONObject jsonResult = null;
        HttpPost method = new HttpPost(url);
        String str = null;
        try {
            // 解决中文乱码问题
            StringEntity entity = new StringEntity(str1, "utf-8");
            entity.setContentEncoding("UTF-8");
            entity.setContentType("application/json");
            method.setEntity(entity);
            if (null != Header.get("Client-Account")) {
                method.setHeader("Client-Account", (String) Header.get("Client-Account"));
            }
            if (null != Header.get("Tenant-Code")) {
                method.setHeader("Tenant-Code", (String) Header.get("Tenant-Code"));
            }
            Gson gson = new Gson();
            System.out.println(gson.toJson(method.getAllHeaders()));
            HttpResponse result = httpClient.execute(method);
            url = URLDecoder.decode(url, "UTF-8");
            /** 请求发送成功，并得到响应 **/
            // System.out.println("EntityUtils.toString(result.getEntity()):" +
            // EntityUtils.toString(result.getEntity()));
            if (result.getStatusLine().getStatusCode() == 200) {
                try {
                    /** 读取服务器返回过来的json字符串数据 **/
                    String s1 = EntityUtils.toString(result.getEntity());
                    //str = new String(s1.getBytes("iso-8859-1"), "utf-8");

                    str = s1;
                } catch (Exception e) {
                    e.printStackTrace();
                }
            } else {
                System.out.println("result.getStatusLine().getStatusCode():" + result.getStatusLine().getStatusCode());
            }
        } catch (IOException e) {
            e.printStackTrace();
            throw e;
        } finally {
            try {
                if (null != method) {
                    method.releaseConnection();
                    method = null;
                }

                closeHttpClient(httpClient);
                httpClient = null;
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
        return str;
    }

    /**
     * 避免HttpClient的”SSLPeerUnverifiedException: peer not authenticated”异常
     * 不用导入SSL证书
     *
     * @param base
     * @return
     */
    public static HttpClient wrapClient(HttpClient base) {
        try {
            SSLContext ctx = SSLContext.getInstance("TLS");
            X509TrustManager tm = new X509TrustManager() {

                @Override
                public void checkClientTrusted(java.security.cert.X509Certificate[] arg0, String arg1)
                        throws java.security.cert.CertificateException {
                    // TODO Auto-generated method stub

                }

                @Override
                public void checkServerTrusted(java.security.cert.X509Certificate[] arg0, String arg1)
                        throws java.security.cert.CertificateException {
                    // TODO Auto-generated method stub

                }

                @Override
                public java.security.cert.X509Certificate[] getAcceptedIssuers() {
                    // TODO Auto-generated method stub
                    return null;
                }
            };
            ctx.init(null, new TrustManager[]{tm}, null);
            SSLConnectionSocketFactory ssf = new SSLConnectionSocketFactory(ctx, NoopHostnameVerifier.INSTANCE);
            CloseableHttpClient httpclient = HttpClients.custom().setSSLSocketFactory(ssf).build();
            return httpclient;
        } catch (Exception ex) {
            ex.printStackTrace();
            return HttpClients.createDefault();
        }
    }

    public static String sendMsgAndRecvByHttps(String url, List<NameValuePair> list) throws Exception {
        CloseableHttpClient httpClient = null;
        HttpPost post = null;
        String result = null;
        String sRet = "-1";
        try {
            httpClient = new SSLClient();
            post = new HttpPost(url);

            // url格式编码
            UrlEncodedFormEntity uefEntity = new UrlEncodedFormEntity(list, "UTF-8");
            post.setEntity(uefEntity);
            System.out.println("HTTPS POST 请求...." + post.getURI());
            // 执行请求
            CloseableHttpResponse httpResponse = httpClient.execute(post);
            try {
                HttpEntity entity = httpResponse.getEntity();
                if (null != entity) {
                    System.out.println("-------------------------------------------------------");
                    sRet = EntityUtils.toString(entity);
                    System.out.println(sRet);
                    System.out.println("-------------------------------------------------------");

                }
            } finally {
                httpResponse.close();
            }

        } catch (Exception e) {
            e.printStackTrace();
            throw e;
        } finally {
            try {
                closeHttpClient(httpClient);
            } catch (Exception e) {
                e.printStackTrace();
            }
        }

        return sRet;
    }

    public static String postJsonByHttps(String url, String str1) throws Exception {

        CloseableHttpClient httpClient = null;
        HttpPost method = null;
        String result = null;
        String sRet = "-1";
        try {
            httpClient = new SSLClient();
            method = new HttpPost(url);

            // url格式编码
            // UrlEncodedFormEntity uefEntity = new UrlEncodedFormEntity(list,
            // "UTF-8");
            // method.setEntity(uefEntity);
            StringEntity entity = new StringEntity(str1, "utf-8");
            entity.setContentEncoding("UTF-8");
            entity.setContentType("application/json");
            method.setEntity(entity);
            System.out.println("HTTPS POST 请求...." + method.getURI());
            // 执行请求
            CloseableHttpResponse httpResponse = httpClient.execute(method);
            try {
                HttpEntity entity1 = httpResponse.getEntity();
                if (null != entity) {
                    System.out.println("-------------------------------------------------------");
                    sRet = EntityUtils.toString(entity1);
                    System.out.println(sRet);
                    System.out.println("-------------------------------------------------------");

                }
            } finally {
                httpResponse.close();
            }

        } catch (Exception e) {
            e.printStackTrace();
            throw e;
        } finally {
            try {
                closeHttpClient(httpClient);
            } catch (Exception e) {
                e.printStackTrace();
            }
        }

        return sRet;
    }

    //post请求

    /**
     * @param url
     * @param headerMap  header 参数
     * @param contentMap body 参数
     * @return
     */
    public static String postMap(String url, Map<String, String> headerMap, Map<String, String> contentMap) {
        String result = null;
        CloseableHttpClient httpClient = HttpClients.createDefault();
        HttpPost post = new HttpPost(url);
        List<NameValuePair> content = new ArrayList<NameValuePair>();
        Iterator iterator = contentMap.entrySet().iterator();           //将content生成entity
        while (iterator.hasNext()) {
            Map.Entry<String, String> elem = (Map.Entry<String, String>) iterator.next();
            content.add(new BasicNameValuePair(elem.getKey(), elem.getValue()));
        }
        CloseableHttpResponse response = null;
        try {
            Iterator headerIterator = headerMap.entrySet().iterator();          //循环增加header
            while (headerIterator.hasNext()) {
                Map.Entry<String, String> elem = (Map.Entry<String, String>) headerIterator.next();
                post.addHeader(elem.getKey(), elem.getValue());
            }
            if (content.size() > 0) {
                UrlEncodedFormEntity entity = new UrlEncodedFormEntity(content, "UTF-8");
                post.setEntity(entity);
            }
            response = httpClient.execute(post);            //发送请求并接收返回数据
            if (response != null && response.getStatusLine().getStatusCode() == 200) {
                HttpEntity entity = response.getEntity();       //获取response的body部分
                result = EntityUtils.toString(entity);          //读取reponse的body部分并转化成字符串
            }
            return result;
        } catch (UnsupportedEncodingException e) {
            e.printStackTrace();
        } catch (ClientProtocolException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        } finally {
            try {
                httpClient.close();
                if (response != null) {
                    response.close();
                }
            } catch (IOException e) {
                e.printStackTrace();
            }

        }
        return null;
    }


    // url格式转map
    public static Map<String, String> paramToMap(String paramStr) {
        String[] params = paramStr.split("&");
        Map<String, String> resMap = new HashMap<String, String>();
        for (int i = 0; i < params.length; i++) {
            String[] param = params[i].split("=");
            if (param.length >= 2) {
                String key = param[0];
                String value = param[1];
                for (int j = 2; j < param.length; j++) {
                    value += "=" + param[j];
                }
                resMap.put(key, value);
            }
        }
        return resMap;
    }

    //Header字符串转map
    public static Map<String, String> splid(String zz) {
        String[] stepOne = zz.split("\n");
        Map<String, String> map = new HashMap<String, String>();
        for (int i = 0; i < stepOne.length; i++) {
            String[] stepTwo = stepOne[i].split(": ");
            if (map.get(stepTwo[0]) == null)
                map.put(stepTwo[0], stepTwo[1]);
            else
                map.put(stepTwo[0], stepTwo[1] + "," + map.get(stepTwo[0]));
        }
        return map;
    }

    /**
     * @param url, params, fileMap
     * @return java.lang.String
     * @throws
     * @author ludexin
     * @Description 发送文件
     * @date 2020/12/11 11:02
     */
    public static String httpClientSendFile(String url, Map<String, Object> params, Map<String, Object> fileMap) throws IOException {
        CloseableHttpClient httpClient = getHttpClient();
        String result = "";
        //每个post参数之间的分隔。随意设定，只要不会和其他的字符串重复即可。
        String boundary = "--------------20200103121104567";
        boundary = "2124efbe-bdba-4bd9-b3a8-8968aa5fe79c";
        // Content-Type: multipart/form-data; boundary=e1475801-88e1-4bef-96f1-c1a942e131bc
        try {
            HttpPost httpPost = new HttpPost(url);
            //设置请求头
            httpPost.setHeader("Content-Type", "multipart/form-data; boundary=" + boundary);

            //HttpEntity builder
            MultipartEntityBuilder builder = MultipartEntityBuilder.create();
            //字符编码
            builder.setCharset(Charset.forName("UTF-8"));
            //模拟浏览器
            builder.setMode(HttpMultipartMode.BROWSER_COMPATIBLE);
            //boundary
            builder.setBoundary(boundary);
            //multipart/form-data
            builder.addPart(String.valueOf(fileMap.get("name")), new FileBody((File) fileMap.get("file")));//相当于<input name='file' type='file'/>
            // binary
//            builder.addBinaryBody("name=\"file\"; filename=\"test.txt\"", new FileInputStream(file), ContentType.MULTIPART_FORM_DATA, file.getName());// 文件流
            //其他参数  上面setCharset()好像没用，得对每一项设置编码格式中文才不会乱码，这里用form-data或者text/plain好像都可以的
            ContentType srcContent = ContentType.create("form-data", Charset.forName("UTF-8"));
            for (Map.Entry<String, Object> entry : params.entrySet()) {
                builder.addTextBody(entry.getKey(), String.valueOf(entry.getValue()), srcContent);
            }
            //HttpEntity
            HttpEntity entity = builder.build();
            httpPost.setEntity(entity);
            // 执行提交
            HttpResponse response = httpClient.execute(httpPost);
            //响应
            HttpEntity responseEntity = response.getEntity();
            if (responseEntity != null) {
                // 将响应内容转换为字符串
                result = EntityUtils.toString(responseEntity, Charset.forName("UTF-8"));
            }
        } catch (Exception e) {
            e.printStackTrace();
            throw e;
        } finally {
            try {
                closeHttpClient(httpClient);
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
        System.err.println("result" + result);
        return result;
    }

    /**
     * http的get请求
     *
     * @param geturl
     * @param param
     * @return
     */
    public static String sendGet(String geturl, String param) {
        HttpURLConnection connection = null;
        InputStream is = null;
        BufferedReader br = null;
        String result = null;// 返回结果字符串
        String httpurl = geturl + "?" + param;
        try {
            // 创建远程url连接对象
            URL url = new URL(httpurl);
            // 通过远程url连接对象打开一个连接，强转成httpURLConnection类
            connection = (HttpURLConnection) url.openConnection();
            // 设置连接方式：get
            connection.setRequestMethod("GET");
            // 设置连接主机服务器的超时时间：15000毫秒
            connection.setConnectTimeout(15000);
            // 设置读取远程返回的数据时间：60000毫秒
            connection.setReadTimeout(60000);
            // 发送请求
            connection.connect();
            // 通过connection连接，获取输入流
            if (connection.getResponseCode() == 200) {
                is = connection.getInputStream();
                // 封装输入流is，并指定字符集
                br = new BufferedReader(new InputStreamReader(is, "UTF-8"));
                // 存放数据
                StringBuffer sbf = new StringBuffer();
                String temp = null;
                while ((temp = br.readLine()) != null) {
                    sbf.append(temp);
                    sbf.append("\r\n");
                }
                result = sbf.toString();
            }
        } catch (Exception e) {

        } finally {
            // 关闭资源
            if (null != br) {
                try {
                    br.close();
                } catch (IOException e) {
                }
            }
            if (null != is) {
                try {
                    is.close();
                } catch (IOException e) {
                }
            }
            connection.disconnect();// 关闭远程连接
        }
        return result;
    }

}
