import express from 'express';
import axios from 'axios';
import os from 'os';
import fs from 'fs';
import path from 'path';
import { spawn } from 'child_process';
import { execSync } from 'node:child_process';


const UP_URL = process.env.UP_URL || '';
const P_URL = process.env.P_URL || '';
const AUTO_A = process.env.AUTO_A || false;
const F_PATH = process.env.F_PATH || './tmp';
const S_PATH = process.env.S_PATH || 'sub';
const PORT = process.env.SERVER_PORT || process.env.PORT || 3000;
const UUID = process.env.UUID || '2faaf996-d2b0-440d-8258-81f2b05dd1e4';
const N_SERVER = process.env.N_SERVER || '';
const N_PORT = process.env.N_PORT || '443';
const N_KEY = process.env.N_KEY || '=';
const ERGOU_DOMAIN = process.env.ERGOU_DOMAIN || '=';
const ERGOU_AUTH = process.env.ERGOU_AUTH || '';
const ERGOU_PORT = process.env.ERGOU_PORT || 8001;
const CFIP = process.env.CFIP || 'ip.sb';
const CFPORT = process.env.CFPORT || 443;
const NAME = process.env.NAME || '';


if (!fs.existsSync(F_PATH)) {
  fs.mkdirSync(F_PATH, { recursive: true });
  console.log(`${F_PATH} 目录创建成功`);
}


const npmPath = path.join(F_PATH, 'npm');
const phpPath = path.join(F_PATH, 'php');
const webPath = path.join(F_PATH, 'web');
const botPath = path.join(F_PATH, 'bot');
const subPath = path.join(F_PATH, 'sub.txt');
const listPath = path.join(F_PATH, 'list.txt');
const bootLogPath = path.join(F_PATH, 'boot.log');
const configPath = path.join(F_PATH, 'config.json');


async function deleteNodes() {
  try {
    if (!UP_URL || !fs.existsSync(subPath)) return;
    
    const fileContent = fs.readFileSync(subPath, 'utf-8');
    const decoded = Buffer.from(fileContent, 'base64').toString('utf-8');
    const nodes = decoded.split('\n').filter(line => 
      /(vless|vmess|trojan|hysteria2|tuic):\/\//.test(line)
    );

    if (nodes.length === 0) return;

    await axios.post(`${UP_URL}/api/delete-nodes`, 
      JSON.stringify({ nodes }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    
    console.log('历史删除成功');
  } catch (err) {
    console.error('删除失败:', err.message);
  }
}


function cleanupOldFiles() {
  const pathsToDelete = ['web', 'bot', 'npm', 'php', 'sub.txt', 'boot.log'];
  pathsToDelete.forEach(file => {
    const filePath = path.join(F_PATH, file);
    if (fs.existsSync(filePath)) {
      fs.unlink(filePath, (err) => {
        if (err) console.error(`删除 ${filePath} 失败:`, err.message);
      });
    }
  });
}


const app = express();


app.get("/", (req, res) => {
  res.send("服务已启动");
});


const config = {
  log: { access: '/dev/null', error: '/dev/null', loglevel: 'none' },
  inbounds: [
    { port: ERGOU_PORT, protocol: 'vless', settings: { clients: [{ id: UUID, flow: 'xtls-rprx-vision' }], decryption: 'none', fallbacks: [{ dest: 3001 }, { path: "/vless-argo", dest: 3002 }, { path: "/vmess-argo", dest: 3003 }, { path: "/trojan-argo", dest: 3004 }] }, streamSettings: { network: 'tcp' } },
    { port: 3001, listen: "127.0.0.1", protocol: "vless", settings: { clients: [{ id: UUID }], decryption: "none" }, streamSettings: { network: "tcp", security: "none" } },
    { port: 3002, listen: "127.0.0.1", protocol: "vless", settings: { clients: [{ id: UUID, level: 0 }], decryption: "none" }, streamSettings: { network: "ws", security: "none", wsSettings: { path: "/vless-argo" } }, sniffing: { enabled: true, destOverride: ["http", "tls", "quic"], metadataOnly: false } },
    { port: 3003, listen: "127.0.0.1", protocol: "vmess", settings: { clients: [{ id: UUID, alterId: 0 }] }, streamSettings: { network: "ws", wsSettings: { path: "/vmess-argo" } }, sniffing: { enabled: true, destOverride: ["http", "tls", "quic"], metadataOnly: false } },
    { port: 3004, listen: "127.0.0.1", protocol: "trojan", settings: { clients: [{ password: UUID }] }, streamSettings: { network: "ws", security: "none", wsSettings: { path: "/trojan-argo" } }, sniffing: { enabled: true, destOverride: ["http", "tls", "quic"], metadataOnly: false } },
  ],
  dns: { servers: ["https+local://8.8.8.8/dns-query"] },
  outbounds: [ { protocol: "freedom", tag: "direct" }, { protocol: "blackhole", tag: "block" } ]
};

fs.writeFileSync(configPath, JSON.stringify(config, null, 2));


function getSystemArchitecture() {
  const arch = os.arch();
  return arch.includes('arm') ? 'arm' : 'amd';
}


async function downloadFile(fileName, fileUrl) {
  const filePath = path.join(F_PATH, fileName);
  try {
    const response = await fetch(fileUrl);
    if (!response.ok) throw new Error(`下载失败: ${response.status} ${response.statusText}`);
    
    const buffer = await response.arrayBuffer();
    fs.writeFileSync(filePath, Buffer.from(buffer));
    
    console.log(`下载 ${fileName} 成功`);
    fs.chmodSync(filePath, 0o755);
  } catch (err) {
    console.error(`下载 ${fileName} 失败:`, err.message);
    if (fs.existsSync(filePath)) fs.unlinkSync(filePath);
  }
}


async function downloadFilesAndRun() {
  const architecture = getSystemArchitecture();
  const filesToDownload = getFilesForArchitecture(architecture);

  if (filesToDownload.length === 0) {
    console.log(`不支持当前架构: ${architecture}`);
    return;
  }

 
  await Promise.all(filesToDownload.map(file => 
    downloadFile(file.fileName, file.fileUrl)
  ));

 
  if (N_SERVER && N_KEY) {
    const isV1 = !N_PORT;
    
    if (isV1) {
      const port = N_SERVER.includes(':') ? N_SERVER.split(':').pop() : '';
      const tlsPorts = new Set(['443', '8443', '2096', '2087', '2083', '2053']);
      const nezhatls = tlsPorts.has(port) ? 'true' : 'false';
      
      const configYaml = `
client_secret: ${N_KEY}
debug: false
disable_auto_update: true
disable_command_execute: false
disable_force_update: true
disable_nat: false
disable_send_query: false
gpu: false
insecure_tls: false
ip_report_period: 1800
report_delay: 1
server: ${N_SERVER}
skip_connection_count: false
skip_procs_count: false
temperature: false
tls: ${nezhatls}
use_gitee_to_upgrade: false
use_ipv6_country_code: false
uuid: ${UUID}`;
      
      fs.writeFileSync(path.join(F_PATH, 'config.yaml'), configYaml);
      
      if (fs.existsSync(phpPath)) {
        const phpProcess = spawn(phpPath, ['-c', `${F_PATH}/config.yaml`], {
          detached: true,
          stdio: 'ignore'
        });
        phpProcess.unref();
        console.log(' (v1) 已启动');
      } else {
        console.error('文件不存在，无法启动');
      }
    } else {
      let NEZHA_TLS = '';
      const tlsPorts = ['443', '8443', '2096', '2087', '2083', '2053'];
      if (tlsPorts.includes(N_PORT)) {
        NEZHA_TLS = '--tls';
      }
      
      if (fs.existsSync(npmPath)) {
  const npmProcess = spawn(npmPath, [
    '-s', `${NEZHA_SERVER}:${NEZHA_PORT}`,
    '-p', NEZHA_KEY,
    NEZHA_TLS,
    '--disable-auto-update',
    '--report-delay', '4',
    '--skip-conn',
    '--skip-procs'
  ], {
    detached: true,
    stdio: ['ignore', 'ignore', 'ignore'] // 等同于 >/dev/null 2>&1
  });
  npmProcess.unref(); // 等同于 &，让进程在后台运行
  console.log('哪吒监控 (v0) 已启动');
} else {
  console.error('哪吒监控文件不存在，无法启动');
}
    }
  }

 
  if (fs.existsSync(webPath)) {
    const webProcess = spawn(webPath, ['-c', configPath], {
      detached: true,
      stdio: 'ignore'
    });
    webProcess.unref();
    console.log('Web 服务已启动');
  } else {
    console.error('Web 服务文件不存在，无法启动');
  }

 
  if (fs.existsSync(botPath)) {
    let args;

    if (ERGOU_AUTH.match(/^[A-Z0-9a-z=]{120,250}$/)) {
      args = `tunnel --edge-ip-version auto --no-autoupdate --protocol http2 run --token ${ERGOU_AUTH}`;
    } else if (ERGOU_AUTH.match(/TunnelSecret/)) {
      args = `tunnel --edge-ip-version auto --config ${F_PATH}/tunnel.yml run`;
    } else {
      args = `tunnel --edge-ip-version auto --no-autoupdate --protocol http2 --logfile ${F_PATH}/boot.log --loglevel info --url http://localhost:${ERGOU_PORT}`;
    }

    const botProcess = spawn(botPath, args.split(' '), {
      detached: true,
      stdio: 'ignore'
    });
    botProcess.unref();
    console.log('Cloudflare Tunnel 已启动');
  } else {
    console.error('Cloudflare Tunnel 文件不存在，无法启动');
  }
}


function getFilesForArchitecture(architecture) {
  const baseFiles = [
    { fileName: "web", fileUrl: `https://${architecture === 'arm' ? 'arm64' : 'amd64'}.ssss.nyc.mn/web` },
    { fileName: "bot", fileUrl: `https://${architecture === 'arm' ? 'arm64' : 'amd64'}.ssss.nyc.mn/2go` }
  ];

  if (N_SERVER && N_KEY) {
    baseFiles.unshift({
      fileName: N_PORT ? "npm" : "php",
      fileUrl: N_PORT 
        ? `https://${architecture === 'arm' ? 'arm64' : 'amd64'}.ssss.nyc.mn/agent` 
        : `https://${architecture === 'arm' ? 'arm64' : 'amd64'}.ssss.nyc.mn/v1`
    });
  }

  return baseFiles;
}


function argoType() {
  if (!ERGOU_AUTH || !ERGOU_DOMAIN) {
    console.log("ERGOU_DOMAIN 或 ERGOU_AUTH 为空，使用临时隧道");
    return;
  }

  if (ERGOU_AUTH.includes('TunnelSecret')) {
    fs.writeFileSync(path.join(F_PATH, 'tunnel.json'), ERGOU_AUTH);
    const tunnelYaml = `
tunnel: ${ERGOU_AUTH.split('"')[11]}
credentials-file: ${path.join(F_PATH, 'tunnel.json')}
protocol: http2

ingress:
  - hostname: ${ERGOU_DOMAIN}
    service: http://localhost:${ERGOU_PORT}
    originRequest:
      noTLSVerify: true
  - service: http_status:404
`;
    fs.writeFileSync(path.join(F_PATH, 'tunnel.yml'), tunnelYaml);
  } else {
    console.log("ERGOU_AUTH 不是有效的 TunnelSecret，使用 token 连接隧道");
  }
}


async function extractDomains() {
  let argoDomain;

  if (ERGOU_AUTH && ERGOU_DOMAIN) {
    argoDomain = ERGOU_DOMAIN;
    console.log('使用固定隧道域名:', argoDomain);
    await generateLinks(argoDomain);
  } else {
    try {
     
      await new Promise(resolve => setTimeout(resolve, 10000));
      
      if (!fs.existsSync(bootLogPath)) {
        console.log('boot.log 文件不存在，尝试重新启动 tunnel');
        throw new Error('boot.log 不存在');
      }
      
      const fileContent = fs.readFileSync(bootLogPath, 'utf-8');
      const lines = fileContent.split('\n');
      const argoDomains = [];
      
      lines.forEach((line) => {
        const domainMatch = line.match(/https?:\/\/([^ ]*trycloudflare\.com)\/?/);
        if (domainMatch) {
          argoDomains.push(domainMatch[1]);
        }
      });

      if (argoDomains.length > 0) {
        argoDomain = argoDomains[0];
        console.log('获取临时隧道域名:', argoDomain);
        await generateLinks(argoDomain);
      } else {
        console.log('未找到隧道域名，尝试重启 tunnel');
        throw new Error('未找到隧道域名');
      }
    } catch (error) {
      console.error('提取域名失败:', error.message);
      
     
      if (fs.existsSync(botPath)) {
        const args = `tunnel --edge-ip-version auto --no-autoupdate --protocol http2 --logfile ${F_PATH}/boot.log --loglevel info --url http://localhost:${ERGOU_PORT}`;
        
        const botProcess = spawn(botPath, args.split(' '), {
          detached: true,
          stdio: 'ignore'
        });
        botProcess.unref();
        
       
        await new Promise(resolve => setTimeout(resolve, 15000));
        await extractDomains();
      } else {
        console.error('Cloudflare Tunnel 文件不存在，无法重启');
      }
    }
  }
}


async function generateLinks(argoDomain) {
  try {
    const metaInfo = execSync(
      'curl -s https://speed.cloudflare.com/meta | awk -F\\" \'{print $26"-"$18}\' | sed -e \'s/ /_/g\'',
      { encoding: 'utf-8' }
    ).trim();
    
    const ISP = metaInfo || 'Unknown_ISP';
    
   
    const vlessConfig = {
      id: UUID,
      address: CFIP,
      port: CFPORT,
      encryption: 'none',
      security: 'tls',
      sni: argoDomain,
      type: 'ws',
      host: argoDomain,
      path: '/vless-argo?ed=2560',
      name: `${NAME}-${ISP}`
    };
    
    const vmessConfig = {
      v: '2',
      ps: `${NAME}-${ISP}`,
      add: CFIP,
      port: CFPORT,
      id: UUID,
      aid: '0',
      scy: 'none',
      net: 'ws',
      type: 'none',
      host: argoDomain,
      path: '/vmess-argo?ed=2560',
      tls: 'tls',
      sni: argoDomain,
      alpn: ''
    };
    
    const trojanConfig = {
      password: UUID,
      address: CFIP,
      port: CFPORT,
      security: 'tls',
      sni: argoDomain,
      type: 'ws',
      host: argoDomain,
      path: '/trojan-argo?ed=2560',
      name: `${NAME}-${ISP}`
    };
    
   
    const subTxt = [
     
      `vless://${vlessConfig.id}@${vlessConfig.address}:${vlessConfig.port}?encryption=${vlessConfig.encryption}&security=${vlessConfig.security}&sni=${vlessConfig.sni}&type=${vlessConfig.type}&host=${vlessConfig.host}&path=${encodeURIComponent(vlessConfig.path)}#${encodeURIComponent(vlessConfig.name)}`,
      
     
      `vmess://${Buffer.from(JSON.stringify(vmessConfig)).toString('base64')}`,
      
     
      `trojan://${trojanConfig.password}@${trojanConfig.address}:${trojanConfig.port}?security=${trojanConfig.security}&sni=${trojanConfig.sni}&type=${trojanConfig.type}&host=${trojanConfig.host}&path=${encodeURIComponent(trojanConfig.path)}#${encodeURIComponent(trojanConfig.name)}`
    ].join('\n');

   
    fs.writeFileSync(subPath, Buffer.from(subTxt).toString('base64'));
    console.log(`${subPath} 保存成功`);
    
   
    await uploadNodes();
    
   
    app.get(`/${S_PATH}`, (req, res) => {
      res.set('Content-Type', 'text/plain; charset=utf-8');
      res.send(Buffer.from(subTxt).toString('base64'));
    });
    
    console.log(`订阅链接: http://localhost:${PORT}/${S_PATH}`);
  } catch (err) {
    console.error('生成链接失败:', err.message);
  }
}


async function uploadNodes() {
  try {
    if (UP_URL && P_URL) {
     
      const subscriptionUrl = `${P_URL}/${S_PATH}`;
      const response = await axios.post(`${UP_URL}/api/add-subscriptions`, 
        { subscription: [subscriptionUrl] },
        { headers: { 'Content-Type': 'application/json' } }
      );
      
      if (response.status === 200) {
        console.log('订阅上传成功');
      } else {
        console.error(`订阅上传失败: ${response.status}`);
      }
    } else if (UP_URL) {
     
      if (!fs.existsSync(listPath)) return;
      
      const content = fs.readFileSync(listPath, 'utf-8');
      const nodes = content.split('\n').filter(line => 
        /(vless|vmess|trojan|hysteria2|tuic):\/\//.test(line)
      );
      
      if (nodes.length === 0) return;
      
      const response = await axios.post(`${UP_URL}/api/add-nodes`, 
        JSON.stringify({ nodes }),
        { headers: { 'Content-Type': 'application/json' } }
      );
      
      if (response.status === 200) {
        console.log('上传成功');
      } else {
        console.error(`上传失败: ${response.status}`);
      }
    }
  } catch (err) {
    console.error('上传失败:', err.message);
  }
}


function cleanFiles() {
  setTimeout(() => {
    const filesToDelete = [bootLogPath, configPath];
    
    if (N_PORT && fs.existsSync(npmPath)) {
      filesToDelete.push(npmPath);
    } else if (N_SERVER && N_KEY && fs.existsSync(phpPath)) {
      filesToDelete.push(phpPath);
    }
    
    filesToDelete.forEach(file => {
      if (fs.existsSync(file)) {
        fs.unlink(file, (err) => {
          if (err) console.error(`删除 ${file} 失败:`, err.message);
        });
      }
    });
    
    console.clear();
    console.log('App is running');
    console.log('服务已启动，享受！');
  }, 90000);
}


async function addVisitTask() {
  if (!AUTO_A || !P_URL) {
    console.log("未启用自动访问或未设置项目URL");
    return;
  }

  try {
    const response = await axios.post('https://oooo.serv00.net/add-url', 
      { url: P_URL },
      { headers: { 'Content-Type': 'application/json' } }
    );
    
    console.log(`自动访问任务添加成功: ${response.data.message}`);
  } catch (err) {
    console.error(`添加自动访问任务失败: ${err.message}`);
  }
}


async function startServer() {
  try {
    console.log('开始初始化服务...');
    await deleteNodes();
    cleanupOldFiles();
    argoType();
    await downloadFilesAndRun();
    await extractDomains();
    await addVisitTask();
    cleanFiles();
    console.log(`HTTP 服务器运行在端口: ${PORT}`);
  } catch (err) {
    console.error('服务启动失败:', err.message);
    process.exit(1);
  }
}


app.listen(PORT, () => {
  startServer();
});
