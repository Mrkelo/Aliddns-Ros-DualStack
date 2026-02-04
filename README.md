# Aliddns-Ros-DualStack

本项目 Fork 自 [lsprain/Aliddns-Ros](https://github.com/lsprain/Aliddns-Ros)。由于原项目已停止维护多年，本项目进行了深度重构与功能增强。

### 🌟 主要改进

* **内核升级**：将阿里 DNS SDK 从第三方包更换为**阿里云官方 SDK**。
* **双栈支持**：原生支持 **IPv4** 与 **IPv6** 解析记录同步更新。
* **架构优化**：重构 RouterOS 脚本，支持单次触发同时更新双栈地址，效率更高。

---

## 一、 部署方式

目前支持 Linux 服务器二进制部署，Docker 镜像正在规划中。

### 1.1 编译与运行

确保你的环境中已安装 Go 1.18 或以上版本。

```bash
# 获取仓库
git clone https://github.com/Mrkelo/Aliddns-Ros-DualStack.git
cd Aliddns-Ros-DualStack

# 编译程序
go build -o aliddns-server main.go

# 测试启动
./aliddns-server

```

### 1.2 使用 Systemd 管理 (推荐)

为了保证服务在后台稳定运行及开机自启，建议创建服务单元文件 `/etc/systemd/system/aliddns.service`:

```ini
[Unit]
Description=AliDDNS Webhook Server for ROS
After=network.target

[Service]
Type=simple
User=root
# 请根据实际路径修改以下两项
WorkingDirectory=/opt/aliddns
ExecStart=/opt/aliddns/aliddns-server
Restart=on-failure

[Install]
WantedBy=multi-user.target

```

**管理命令：**

```bash
systemctl daemon-reload
systemctl enable aliddns
systemctl start aliddns

```

---

## 二、 RouterOS 7.x 脚本配置

在 ROS 的 `System -> Scripts` 中添加以下脚本。请根据注释修改对应的**密钥**和**接口名称**。

```routeros
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

# ======= IPv6 设置 =======
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

---

## 三、 API 接口说明

如果你希望通过其他工具（如 `curl`）调用，接口定义如下：

* **Method**: `GET` / `POST`
* **URL**: `http://{IP}:8800/aliddns`
* **参数说明**:

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| AccessKeyID | 是 | 阿里云 RAM 账号 Key |
| AccessKeySecret | 是 | 阿里云 RAM 账号 Secret |
| DomainName | 是 | 主域名 (例: `baidu.com`) |
| RR | 是 | 主机记录 (例: `www` 或 `home`) |
| IpAddr | 是 | 需要指向的 IP 地址 |

---

## ⚠️ 安全警告

> [!CAUTION]
> **请勿公网暴露此服务！**
> 由于 ROS 脚本限制，目前 AccessKey 采用明文传输。为了您的账号安全：
> 1. 请务必将本项目部署在**内网**环境。
> 2. 建议在阿里云控制台为 AccessKey 配置**最小权限原则**（仅授予云解析权限）。
> 3. 本项目不提供、也不建议使用任何公共服务接口。
> 
> 

---

## 鸣谢

感谢原作者 [lsprain](https://github.com/lsprain) 的灵感与初始代码贡献。

---
