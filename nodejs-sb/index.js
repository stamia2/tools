const express = require("express");
const app = express();
const axios = require("axios");
const os = require('os');
const fs = require("fs");
const path = require("path");
const { promisify } = require('util');
const exec = promisify(require('child_process').exec);
const { execSync } = require('child_process');        
const UP_URL = process.env.UP_URL || '';      
const P_URL = process.env.P_URL || '';    
const AUTO_A = process.env.AUTO_A || false; 
const F_PATH = process.env.F_PATH || './tmp';   
const S_PATH = process.env.S_PATH || 'sss';       
const PORT = process.env.SERVER_PORT || process.env.PORT || 3000;        
const UUID = process.env.UUID || '9afd1229-b893-40c1-84dd-51e7ce204913'; 
const N_SERVER = process.env.N_SERVER || '';        
const N_PORT = process.env.N_PORT || '';            
const N_KEY = process.env.N_KEY || '';              
const ERGOU_DOMAIN = process.env.ERGOU_DOMAIN || '';          
const ERGOU_AUTH = process.env.ERGOU_AUTH || '';              
const ERGOU_PORT = process.env.ERGOU_PORT || 8001;            
const CFIP = process.env.CFIP || 'ip.sb';         
const CFPORT = process.env.CFPORT || 443;                   
const NAME = process.env.NAME || 'Vls';                     


if (!fs.existsSync(F_PATH)) {
  fs.mkdirSync(F_PATH);
  console.log(`${F_PATH} is created`);
} else {
  console.log(`${F_PATH} already exists`);
}

let npmPath = path.join(F_PATH, 'npm');
let phpPath = path.join(F_PATH, 'php');
let webPath = path.join(F_PATH, 'web');
let botPath = path.join(F_PATH, 'bot');
let subPath = path.join(F_PATH, 'sub.txt');
let listPath = path.join(F_PATH, 'list.txt');
let bootLogPath = path.join(F_PATH, 'boot.log');
let configPath = path.join(F_PATH, 'config.json');


function deleteNodes() {
  try {
    if (!UP_URL) return;
    if (!fs.existsSync(subPath)) return;

    let fileContent;
    try {
      fileContent = fs.readFileSync(subPath, 'utf-8');
    } catch {
      return null;
    }

    const decoded = Buffer.from(fileContent, 'base64').toString('utf-8');
    const nodes = decoded.split('\n').filter(line => 
      /(vless|vmess|trojan|hysteria2|tuic):\/\
    );

    if (nodes.length === 0) return;

    return axios.post(`${UP_URL}/api/delete-nodes`, 
      JSON.stringify({ nodes }),
      { headers: { 'Content-Type': 'application/json' } }
    ).catch((error) => { 
      return null; 
    });
  } catch (err) {
    return null;
  }
}


function cleanupOldFiles() {
  const pathsToDelete = ['web', 'bot', 'npm', 'php', 'sub.txt', 'boot.log'];
  pathsToDelete.forEach(file => {
    const filePath = path.join(F_PATH, file);
    fs.unlink(filePath, () => {});
  });
}


app.get("/", function(req, res) {
  res.send("Hello world!");
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
  dns: { servers: ["https+local:
  outbounds: [ { protocol: "freedom", tag: "direct" }, {protocol: "blackhole", tag: "block"} ]
};
fs.writeFileSync(path.join(F_PATH, 'config.json'), JSON.stringify(config, null, 2));


function getSystemArchitecture() {
  const arch = os.arch();
  if (arch === 'arm' || arch === 'arm64' || arch === 'aarch64') {
    return 'arm';
  } else {
    return 'amd';
  }
}


function downloadFile(fileName, fileUrl, callback) {
  const filePath = path.join(F_PATH, fileName);
  const writer = fs.createWriteStream(filePath);

  axios({
    method: 'get',
    url: fileUrl,
    responseType: 'stream',
  })
    .then(response => {
      response.data.pipe(writer);

      writer.on('finish', () => {
        writer.close();
        console.log(`Download ${fileName} successfully`);
        callback(null, fileName);
      });

      writer.on('error', err => {
        fs.unlink(filePath, () => { });
        const errorMessage = `Download ${fileName} failed: ${err.message}`;
        console.error(errorMessage); 
        callback(errorMessage);
      });
    })
    .catch(err => {
      const errorMessage = `Download ${fileName} failed: ${err.message}`;
      console.error(errorMessage); 
      callback(errorMessage);
    });
}


async function downloadFilesAndRun() {
  const architecture = getSystemArchitecture();
  const filesToDownload = getFilesForArchitecture(architecture);

  if (filesToDownload.length === 0) {
    console.log(`Can't find a file for the current architecture`);
    return;
  }

  const downloadPromises = filesToDownload.map(fileInfo => {
    return new Promise((resolve, reject) => {
      downloadFile(fileInfo.fileName, fileInfo.fileUrl, (err, fileName) => {
        if (err) {
          reject(err);
        } else {
          resolve(fileName);
        }
      });
    });
  });

  try {
    await Promise.all(downloadPromises);
  } catch (err) {
    console.error('Error downloading files:', err);
    return;
  }
  
  function authorizeFiles(filePaths) {
    const newPermissions = 0o775;
    filePaths.forEach(relativeFilePath => {
      const absoluteFilePath = path.join(F_PATH, relativeFilePath);
      if (fs.existsSync(absoluteFilePath)) {
        fs.chmod(absoluteFilePath, newPermissions, (err) => {
          if (err) {
            console.error(`Empowerment failed for ${absoluteFilePath}: ${err}`);
          } else {
            console.log(`Empowerment success for ${absoluteFilePath}: ${newPermissions.toString(8)}`);
          }
        });
      }
    });
  }
  const filesToAuthorize = N_PORT ? ['./npm', './web', './bot'] : ['./php', './web', './bot'];
  authorizeFiles(filesToAuthorize);

 
  if (N_SERVER && N_KEY) {
    if (!N_PORT) {
      
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
      
      
      const command = `nohup ${F_PATH}/php -c "${F_PATH}/config.yaml" >/dev/null 2>&1 &`;
      try {
        await exec(command);
        console.log('php is running');
        await new Promise((resolve) => setTimeout(resolve, 1000));
      } catch (error) {
        console.error(`php running error: ${error}`);
      }
    } else {
      let NEZHA_TLS = '';
      const tlsPorts = ['443', '8443', '2096', '2087', '2083', '2053'];
      if (tlsPorts.includes(N_PORT)) {
        NEZHA_TLS = '--tls';
      }
      const command = `nohup ${F_PATH}/npm -s ${N_SERVER}:${N_PORT} -p ${N_KEY} ${NEZHA_TLS} >/dev/null 2>&1 &`;
      try {
        await exec(command);
        console.log('npm is running');
        await new Promise((resolve) => setTimeout(resolve, 1000));
      } catch (error) {
        console.error(`npm running error: ${error}`);
      }
    }
  } else {
    console.log('NEZHA variable is empty,skip running');
  }
 
  const command1 = `nohup ${F_PATH}/web -c ${F_PATH}/config.json >/dev/null 2>&1 &`;
  try {
    await exec(command1);
    console.log('web is running');
    await new Promise((resolve) => setTimeout(resolve, 1000));
  } catch (error) {
    console.error(`web running error: ${error}`);
  }

  
  if (fs.existsSync(path.join(F_PATH, 'bot'))) {
    let args;

    if (ERGOU_AUTH.match(/^[A-Z0-9a-z=]{120,250}$/)) {
      args = `tunnel --edge-ip-version auto --no-autoupdate --protocol http2 run --token ${ERGOU_AUTH}`;
    } else if (ERGOU_AUTH.match(/TunnelSecret/)) {
      args = `tunnel --edge-ip-version auto --config ${F_PATH}/tunnel.yml run`;
    } else {
      args = `tunnel --edge-ip-version auto --no-autoupdate --protocol http2 --logfile ${F_PATH}/boot.log --loglevel info --url http:
    }

    try {
      await exec(`nohup ${F_PATH}/bot ${args} >/dev/null 2>&1 &`);
      console.log('bot is running');
      await new Promise((resolve) => setTimeout(resolve, 2000));
    } catch (error) {
      console.error(`Error executing command: ${error}`);
    }
  }
  await new Promise((resolve) => setTimeout(resolve, 5000));

}


function getFilesForArchitecture(architecture) {
  let baseFiles;
  if (architecture === 'arm') {
    baseFiles = [
      { fileName: "web", fileUrl: "https:
      { fileName: "bot", fileUrl: "https:
    ];
  } else {
    baseFiles = [
      { fileName: "web", fileUrl: "https:
      { fileName: "bot", fileUrl: "https:
    ];
  }

  if (N_SERVER && N_KEY) {
    if (N_PORT) {
      const npmUrl = architecture === 'arm' 
        ? "https:
        : "https:
        baseFiles.unshift({ 
          fileName: "npm", 
          fileUrl: npmUrl 
        });
    } else {
      const phpUrl = architecture === 'arm' 
        ? "https:
        : "https:
      baseFiles.unshift({ 
        fileName: "php", 
        fileUrl: phpUrl
      });
    }
  }

  return baseFiles;
}


function argoType() {
  if (!ERGOU_AUTH || !ERGOU_DOMAIN) {
    console.log("ERGOU_DOMAIN or ERGOU_AUTH variable is empty, use quick tunnels");
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
      service: http:
      originRequest:
        noTLSVerify: true
    - service: http_status:404
  `;
    fs.writeFileSync(path.join(F_PATH, 'tunnel.yml'), tunnelYaml);
  } else {
    console.log("ERGOU_AUTH mismatch TunnelSecret,use token connect to tunnel");
  }
}
argoType();


async function extractDomains() {
  let argoDomain;

  if (ERGOU_AUTH && ERGOU_DOMAIN) {
    argoDomain = ERGOU_DOMAIN;
    console.log('ERGOU_DOMAIN:', argoDomain);
    await generateLinks(argoDomain);
  } else {
    try {
      const fileContent = fs.readFileSync(path.join(F_PATH, 'boot.log'), 'utf-8');
      const lines = fileContent.split('\n');
      const argoDomains = [];
      lines.forEach((line) => {
        const domainMatch = line.match(/https?:\/\/([^ ]*trycloudflare\.com)\/?/);
        if (domainMatch) {
          const domain = domainMatch[1];
          argoDomains.push(domain);
        }
      });

      if (argoDomains.length > 0) {
        argoDomain = argoDomains[0];
        console.log('ArgoDomain:', argoDomain);
        await generateLinks(argoDomain);
      } else {
        console.log('ArgoDomain not found, re-running bot to obtain ArgoDomain');
        
        fs.unlinkSync(path.join(F_PATH, 'boot.log'));
        async function killBotProcess() {
          try {
            await exec('pkill -f "[b]ot" > /dev/null 2>&1');
          } catch (error) {
            
          }
        }
        killBotProcess();
        await new Promise((resolve) => setTimeout(resolve, 3000));
        const args = `tunnel --edge-ip-version auto --no-autoupdate --protocol http2 --logfile ${F_PATH}/boot.log --loglevel info --url http:
        try {
          await exec(`nohup ${path.join(F_PATH, 'bot')} ${args} >/dev/null 2>&1 &`);
          console.log('bot is running.');
          await new Promise((resolve) => setTimeout(resolve, 3000));
          await extractDomains(); 
        } catch (error) {
          console.error(`Error executing command: ${error}`);
        }
      }
    } catch (error) {
      console.error('Error reading boot.log:', error);
    }
  }

  
  async function generateLinks(argoDomain) {
    const metaInfo = execSync(
      'curl -s https:
      { encoding: 'utf-8' }
    );
    const ISP = metaInfo.trim();

    return new Promise((resolve) => {
      setTimeout(() => {
        const VMESS = { v: '2', ps: `${NAME}-${ISP}`, add: CFIP, port: CFPORT, id: UUID, aid: '0', scy: 'none', net: 'ws', type: 'none', host: argoDomain, path: '/vmess-argo?ed=2560', tls: 'tls', sni: argoDomain, alpn: '' };
        const subTxt = `
vless:
  
vmess:
  
trojan:
    `;
        
        console.log(Buffer.from(subTxt).toString('base64'));
        fs.writeFileSync(subPath, Buffer.from(subTxt).toString('base64'));
        console.log(`${F_PATH}/sub.txt saved successfully`);
        uplodNodes();
        
        app.get(`/${S_PATH}`, (req, res) => {
          const encodedContent = Buffer.from(subTxt).toString('base64');
          res.set('Content-Type', 'text/plain; charset=utf-8');
          res.send(encodedContent);
        });
        resolve(subTxt);
      }, 2000);
    });
  }
}


async function uplodNodes() {
  if (UP_URL && P_URL) {
    const subscriptionUrl = `${P_URL}/${S_PATH}`;
    const jsonData = {
      subscription: [subscriptionUrl]
    };
    try {
        const response = await axios.post(`${UP_URL}/api/add-subscriptions`, jsonData, {
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (response.status === 200) {
            console.log('Subscription uploaded successfully');
        } else {
          return null;
          
        }
    } catch (error) {
        if (error.response) {
            if (error.response.status === 400) {
              
            }
        }
    }
  } else if (UP_URL) {
      if (!fs.existsSync(listPath)) return;
      const content = fs.readFileSync(listPath, 'utf-8');
      const nodes = content.split('\n').filter(line => /(vless|vmess|trojan|hysteria2|tuic):\/\

      if (nodes.length === 0) return;

      const jsonData = JSON.stringify({ nodes });

      try {
          await axios.post(`${UP_URL}/api/add-nodes`, jsonData, {
              headers: { 'Content-Type': 'application/json' }
          });
          if (response.status === 200) {
            console.log('Subscription uploaded successfully');
        } else {
            return null;
        }
      } catch (error) {
          return null;
      }
  } else {
      
      return;
  }
}


function cleanFiles() {
  setTimeout(() => {
    const filesToDelete = [bootLogPath, configPath, webPath, botPath, phpPath, npmPath];  
    
    if (N_PORT) {
      filesToDelete.push(npmPath);
    } else if (N_SERVER && N_KEY) {
      filesToDelete.push(phpPath);
    }

    exec(`rm -rf ${filesToDelete.join(' ')} >/dev/null 2>&1`, (error) => {
      console.clear();
      console.log('App is running');
      console.log('Thank you for using this script, enjoy!');
    });
  }, 90000); 
}
cleanFiles();


async function AddVisitTask() {
  if (!AUTO_A || !P_URL) {
    console.log("Skipping adding automatic access task");
    return;
  }

  try {
    const response = await axios.post('https:
      url: P_URL
    }, {
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    console.log(`automatic access task added successfully`);
  } catch (error) {
    console.error(`添加URL失败: ${error.message}`);
  }
}


async function startserver() {
  deleteNodes();
  cleanupOldFiles();
  await downloadFilesAndRun();
  await extractDomains();
  AddVisitTask();
}
startserver();

app.listen(PORT, () => console.log(`http server is running on port:${PORT}!`));
