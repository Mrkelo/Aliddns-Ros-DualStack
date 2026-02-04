# 一、部署方式
目前仅支持服务器部署方式
---
## 1.1、服务器部署
可使用github上的release中的二进制文件，也可以自己编译
### 二进制程序编译
```bash
git pull
cd Aliddns-Ros-DualStack
go build -o aliddns-server
```

### 服务启动
**直接运行：**
```bash
./aliddns-server
```

**直接后台运行：**
```bash
nohup ./aliddns-server > aliddns.log 2>&1 &
```

### 使用Systemd 服务 (推荐)
创建文件 `/etc/systemd/system/aliddns.service`:
```ini
[Unit]
Description=AliDDNS Webhook Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/path/to/directory
ExecStart=/path/to/directory/aliddns-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```
启动服务：
```bash
systemctl enable aliddns
systemctl start aliddns
```
## 1.2、Docker容器部署（本fork暂无docker）

---
# 二、RouterOS7.x 脚本代码
ROS路由脚本如下： 
请修改脚本中的 AccessKeyID、AccessKeySecret、DomainName、pppoe、v6Interface、v6Pool、服务地址 参数后使用
```
# ======= 基础账号配置 =======
:local AccessKeyID "xxxx"
:local AccessKeySecret "xxxx"
:local DomainName "testddns.xxxx"
# PPPOE 接口名称用于 IPv4
:local pppoe "pppoe-out1"

# ======= IPv4 设置 =======
:local RR4 "home"
:local IpAddr4 [/ip address get [/ip address find interface=$pppoe] address]
:set IpAddr4 [:pick $IpAddr4 0 [:find $IpAddr4 "/"]]

# ======= IPv6 设置 (根据你测试成功的参数) =======
:local RR6 "home6"
:local v6Interface "lan1"
:local v6Pool "ipv6_cu"
:local IpAddr6 ""

# 使用 foreach 确保兼容性，获取后立即截断掩码
:foreach i in=[/ipv6 address find interface=$v6Interface from-pool=$v6Pool] do={
    :local rawAddr [/ipv6 address get $i address]
    :set IpAddr6 [:pick $rawAddr 0 [:find $rawAddr "/"]]
}

# ======= 执行更新逻辑 =======
:log info "DDNS: IPv4=$IpAddr4, IPv6=$IpAddr6"

# 更新 IPv4
:if ([:len $IpAddr4] > 0) do={
    :local url4 "http://服务地址:8800/aliddns?AccessKeyID=$AccessKeyID&AccessKeySecret=$AccessKeySecret&RR=$RR4&DomainName=$DomainName&IpAddr=$IpAddr4"
    :do {
        /tool fetch url=$url4 mode=http http-method=get keep-result=no
        :log info "IPv4 DDNS 请求已发送"
    } on-error={ :log error "IPv4 DDNS 访问失败" }
}

# 更新 IPv6
:if ([:len $IpAddr6] > 0) do={
    :local url6 "http://服务地址:8800/aliddns?AccessKeyID=$AccessKeyID&AccessKeySecret=$AccessKeySecret&RR=$RR6&DomainName=$DomainName&IpAddr=$IpAddr6"
    :do {
        /tool fetch url=$url6 mode=http http-method=get keep-result=no
        :log info "IPv6 DDNS 请求已发送"
    } on-error={ :log error "IPv6 DDNS 访问失败" }
}
```
# 三、其它方式
method：```post```   
url：```http://服务地址:8800/aliddns?AccessKeyID=XXXXXX&AccessKeySecret=XXXXXX&RR=XX&DomainName=XXX&IpAddr=XXX```

# 四、公共服务接口
因涉及明文传输用户阿里云的AccessKeyID和AccessKeySecret，存在安全风险，本项目暂不提供公共服务接口，请务必自行部署与内网后使用。

