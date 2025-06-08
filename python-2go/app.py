import os
import re
import shutil
import subprocess
import http.server
import socketserver
import threading
import requests
from flask import Flask
import json
import time
import base64

app = Flask(__name__)


F_PATH = os.environ.get('F_PATH', './tmp')
P_URL = os.environ.get('URL', '') 
INTERVAL_SECONDS = int(os.environ.get("TIME", 120))                         
UUID = os.environ.get('UUID', '0004add9-5c68-8bab-870c-08cd5320df00')       
N_SERVER = os.environ.get('N_SERVER', 'nz.f4i.cn')                  
N_PORT = os.environ.get('N_PORT', '5555')                           
N_KEY = os.environ.get('N_KEY', '')                                 
ERGOU_DOMAIN = os.environ.get('ERGOU_DOMAIN', '')                             
ERGOU_AUTH = os.environ.get('ERGOU_AUTH', '')                                 
ERGOU_PORT = int(os.environ.get('ERGOU_PORT', 8001))                          
CFIP = os.environ.get('CFIP', 'www.visa.com.tw')                            
CFPORT = int(os.environ.get('CFPORT', 443))                                 
NAME = os.environ.get('NAME', 'Vls')                                        
PORT = int(os.environ.get('SERVER_PORT') or os.environ.get('PORT') or 3000) 


if not os.path.exists(F_PATH):
    os.makedirs(F_PATH)
    print(f"{F_PATH} has been created")
else:
    print(f"{F_PATH} already exists")


paths_to_delete = ['boot.log', 'list.txt','sub.txt', 'npm', 'web', 'bot', 'tunnel.yml', 'tunnel.json']
for file in paths_to_delete:
    F_PATH = os.path.join(F_PATH, file)
    try:
        os.unlink(F_PATH)
        print(f"{F_PATH} has been deleted")
    except Exception as e:
        print(f"Skip Delete {F_PATH}")


class MyHandler(http.server.SimpleHTTPRequestHandler):

    def log_message(self, format, *args):
        pass

    def do_GET(self):
        if self.path == '/':
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b'Hello, world')
        elif self.path == '/sub':
            try:
                with open(os.path.join(F_PATH, 'sub.txt'), 'rb') as file:
                    content = file.read()
                self.send_response(200)
                self.send_header('Content-Type', 'text/plain; charset=utf-8')
                self.end_headers()
                self.wfile.write(content)
            except FileNotFoundError:
                self.send_response(500)
                self.end_headers()
                self.wfile.write(b'Error reading file')
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b'Not found')

httpd = socketserver.TCPServer(('', PORT), MyHandler)
server_thread = threading.Thread(target=httpd.serve_forever)
server_thread.daemon = True
server_thread.start()


def generate_config():
    config ={"log":{"access":"/dev/null","error":"/dev/null","loglevel":"none",},"inbounds":[{"port":ERGOU_PORT ,"protocol":"vless","settings":{"clients":[{"id":UUID ,"flow":"xtls-rprx-vision",},],"decryption":"none","fallbacks":[{"dest":3001 },{"path":"/vless-argo","dest":3002 },{"path":"/vmess-argo","dest":3003 },{"path":"/trojan-argo","dest":3004 },],},"streamSettings":{"network":"tcp",},},{"port":3001 ,"listen":"127.0.0.1","protocol":"vless","settings":{"clients":[{"id":UUID },],"decryption":"none"},"streamSettings":{"network":"ws","security":"none"}},{"port":3002 ,"listen":"127.0.0.1","protocol":"vless","settings":{"clients":[{"id":UUID ,"level":0 }],"decryption":"none"},"streamSettings":{"network":"ws","security":"none","wsSettings":{"path":"/vless-argo"}},"sniffing":{"enabled":True ,"destOverride":["http","tls","quic"],"metadataOnly":False }},{"port":3003 ,"listen":"127.0.0.1","protocol":"vmess","settings":{"clients":[{"id":UUID ,"alterId":0 }]},"streamSettings":{"network":"ws","wsSettings":{"path":"/vmess-argo"}},"sniffing":{"enabled":True ,"destOverride":["http","tls","quic"],"metadataOnly":False }},{"port":3004 ,"listen":"127.0.0.1","protocol":"trojan","settings":{"clients":[{"password":UUID },]},"streamSettings":{"network":"ws","security":"none","wsSettings":{"path":"/trojan-argo"}},"sniffing":{"enabled":True ,"destOverride":["http","tls","quic"],"metadataOnly":False }},],"dns":{"servers":["https+local://8.8.8.8/dns-query"]},"outbounds":[{"protocol":"freedom","tag": "direct" },{"protocol":"blackhole","tag":"block"}]}
    with open(os.path.join(F_PATH, 'config.json'), 'w', encoding='utf-8') as config_file:
        json.dump(config, config_file, ensure_ascii=False, indent=2)

generate_config()


def get_system_architecture():
    arch = os.uname().machine
    if 'arm' in arch or 'aarch64' in arch or 'arm64' in arch:
        return 'arm'
    else:
        return 'amd'


def download_file(file_name, file_url):
    F_PATH = os.path.join(F_PATH, file_name)
    with requests.get(file_url, stream=True) as response, open(F_PATH, 'wb') as file:
        shutil.copyfileobj(response.raw, file)


def download_files_and_run():
    architecture = get_system_architecture()
    files_to_download = get_files_for_architecture(architecture)

    if not files_to_download:
        print("Can't find a file for the current architecture")
        return

    for file_info in files_to_download:
        try:
            download_file(file_info['file_name'], file_info['file_url'])
            print(f"Downloaded {file_info['file_name']} successfully")
        except Exception as e:
            print(f"Download {file_info['file_name']} failed: {e}")

    
    files_to_authorize = ['npm', 'web', 'bot']
    authorize_files(files_to_authorize)

    
    N_TLS = ''
    valid_ports = ['443', '8443', '2096', '2087', '2083', '2053']
    if N_SERVER and N_PORT and N_KEY:
        if N_PORT in valid_ports:
          N_TLS = '--tls'
        command = f"nohup {F_PATH}/npm -s {N_SERVER}:{N_PORT} -p {N_KEY} {N_TLS} --disable-auto-update --report-delay 4 --skip-conn --skip-procs >/dev/null 2>&1 &"
        try:
            subprocess.run(command, shell=True, check=True)
            print('npm is running')
            subprocess.run('sleep 1', shell=True)  
        except subprocess.CalledProcessError as e:
            print(f'npm running error: {e}')
    else:
        print('N variable is empty, skip running')

    
    command1 = f"nohup {F_PATH}/web -c {F_PATH}/config.json >/dev/null 2>&1 &"
    try:
        subprocess.run(command1, shell=True, check=True)
        print('web is running')
        subprocess.run('sleep 1', shell=True)  
    except subprocess.CalledProcessError as e:
        print(f'web running error: {e}')

    
    if os.path.exists(os.path.join(F_PATH, 'bot')):
		
        args = get_cloud_flare_args()
        
        try:
            subprocess.run(f"nohup {F_PATH}/bot {args} >/dev/null 2>&1 &", shell=True, check=True)
            print('bot is running')
            subprocess.run('sleep 2', shell=True)  
        except subprocess.CalledProcessError as e:
            print(f'Error executing command: {e}')

    subprocess.run('sleep 3', shell=True)  
	
   
def get_cloud_flare_args():
    
    processed_auth = ERGOU_AUTH
    try:
        auth_data = json.loads(ERGOU_AUTH)
        if 'TunnelSecret' in auth_data and 'AccountTag' in auth_data and 'TunnelID' in auth_data:
            processed_auth = 'TunnelSecret'
    except json.JSONDecodeError:
        pass

    
    if not processed_auth and not ERGOU_DOMAIN:
        args = f'tunnel --edge-ip-version auto --no-autoupdate --protocol http2 --logfile {F_PATH}/boot.log --loglevel info --url http://localhost:{ERGOU_PORT}'
    elif processed_auth == 'TunnelSecret':
        args = f'tunnel --edge-ip-version auto --config {F_PATH}/tunnel.yml run'
    elif processed_auth and ERGOU_DOMAIN and 120 <= len(processed_auth) <= 250:
        args = f'tunnel --edge-ip-version auto --no-autoupdate --protocol http2 run --token {processed_auth}'
    else:
        
        args = f'tunnel --edge-ip-version auto --no-autoupdate --protocol http2 --logfile {F_PATH}/boot.log --loglevel info --url http://localhost:{ERGOU_PORT}'

    return args


def get_files_for_architecture(architecture):
    if architecture == 'arm':
        return [
            {'file_name': 'npm', 'file_url': 'https://arm64.ssss.nyc.mn/agent'},
            {'file_name': 'web', 'file_url': 'https://arm64.ssss.nyc.mn/web'},
            {'file_name': 'bot', 'file_url': 'https://arm64.ssss.nyc.mn/2go'},
        ]
    elif architecture == 'amd':
        return [
            {'file_name': 'npm', 'file_url': 'https://amd64.ssss.nyc.mn/agent'},
            {'file_name': 'web', 'file_url': 'https://amd64.ssss.nyc.mn/web'},
            {'file_name': 'bot', 'file_url': 'https://amd64.ssss.nyc.mn/2go'},
        ]
    return []


def authorize_files(F_PATHs):
    new_permissions = 0o775

    for relative_F_PATH in F_PATHs:
        absolute_F_PATH = os.path.join(F_PATH, relative_F_PATH)
        try:
            os.chmod(absolute_F_PATH, new_permissions)
            print(f"Empowerment success for {absolute_F_PATH}: {oct(new_permissions)}")
        except Exception as e:
            print(f"Empowerment failed for {absolute_F_PATH}: {e}")



def argo_config():
    if not ERGOU_AUTH or not ERGOU_DOMAIN:
        print("ERGOU_DOMAIN or ERGOU_AUTH is empty, use quick Tunnels")
        return

    if 'TunnelSecret' in ERGOU_AUTH:
        with open(os.path.join(F_PATH, 'tunnel.json'), 'w') as file:
            file.write(ERGOU_AUTH)
        tunnel_yaml = f"""
tunnel: {ERGOU_AUTH.split('"')[11]}
credentials-file: {os.path.join(F_PATH, 'tunnel.json')}
protocol: http2

ingress:
  - hostname: {ERGOU_DOMAIN}
    service: http://localhost:{ERGOU_PORT}
    originRequest:
      noTLSVerify: true
  - service: http_status:404
  """
        with open(os.path.join(F_PATH, 'tunnel.yml'), 'w') as file:
            file.write(tunnel_yaml)
    else:
        print("Use token connect to tunnel")

argo_config()


def extract_domains():
    ERGOU_DOMAIN = ''

    if ERGOU_AUTH and ERGOU_DOMAIN:
        ERGOU_DOMAIN = ERGOU_DOMAIN
        print('ERGOU_DOMAIN:', ERGOU_DOMAIN)
        generate_links(ERGOU_DOMAIN)
    else:
        try:
            with open(os.path.join(F_PATH, 'boot.log'), 'r', encoding='utf-8') as file:
                content = file.read()
                
                match = re.search(r'https://([^ ]+\.trycloudflare\.com)', content)
                if match:
                    ERGOU_DOMAIN = match.group(1)
                    print('ArgoDomain:', ERGOU_DOMAIN)
                    generate_links(ERGOU_DOMAIN)
                else:
                    print('ArgoDomain not found, re-running bot to obtain ArgoDomain')
                    
                    try:
                        subprocess.run("pkill -f 'bot tunnel'", shell=True)
                        print('Stopped existing bot process')
                    except Exception as e:
                        print(f'Error stopping bot process: {e}')
                    
                    time.sleep(2)  
                    
                    os.remove(os.path.join(F_PATH, 'boot.log'))
                    
                    
                    max_retries = 10
                    for attempt in range(max_retries):
                        print(f'Attempt {attempt + 1} of {max_retries}')
                        args = f"tunnel --edge-ip-version auto --no-autoupdate --protocol http2 --logfile {F_PATH}/boot.log --loglevel info --url http://localhost:{ERGOU_PORT}"
                        try:
                            subprocess.run(f"nohup {F_PATH}/bot {args} >/dev/null 2>&1 &", shell=True, check=True)
                            print('bot is running')
                            time.sleep(3)
                            
                            with open(os.path.join(F_PATH, 'boot.log'), 'r', encoding='utf-8') as file:
                                content = file.read()
                                match = re.search(r'https://([^ ]+\.trycloudflare\.com)', content)
                                if match:
                                    ERGOU_DOMAIN = match.group(1)
                                    print('ArgoDomain:', ERGOU_DOMAIN)
                                    generate_links(ERGOU_DOMAIN)
                                    break
                            if attempt < max_retries - 1:
                                print('ArgoDomain not found, retrying...')
                                subprocess.run("pkill -f 'bot tunnel'", shell=True)
                                time.sleep(2)
                        except subprocess.CalledProcessError as e:
                            print(f"Error executing command: {e}")
                        except Exception as e:
                            print(f"Error: {e}")
                    else:  
                        print("Failed to obtain ArgoDomain after maximum retries")
        except IndexError as e:
            print(f"IndexError while reading boot.log: {e}")
        except Exception as e:
            print(f"Error reading boot.log: {e}")



def generate_links(ERGOU_DOMAIN):
    meta_info = subprocess.run(['curl', '-s', 'https://speed.cloudflare.com/meta'], capture_output=True, text=True)
    meta_info = meta_info.stdout.split('"')
    ISP = f"{meta_info[25]}-{meta_info[17]}".replace(' ', '_').strip()

    time.sleep(2)
    VMESS = {"v": "2", "ps": f"{NAME}-{ISP}", "add": CFIP, "port": CFPORT, "id": UUID, "aid": "0", "scy": "none", "net": "ws", "type": "none", "host": ERGOU_DOMAIN, "path": "/vmess-argo?ed=2048", "tls": "tls", "sni": ERGOU_DOMAIN, "alpn": ""}
 
    list_txt = f"""
vless://{UUID}@{CFIP}:{CFPORT}?encryption=none&security=tls&sni={ERGOU_DOMAIN}&type=ws&host={ERGOU_DOMAIN}&path=%2Fvless-argo%3Fed%3D2048#{NAME}-{ISP}
  
vmess://{ base64.b64encode(json.dumps(VMESS).encode('utf-8')).decode('utf-8')}

trojan://{UUID}@{CFIP}:{CFPORT}?security=tls&sni={ERGOU_DOMAIN}&type=ws&host={ERGOU_DOMAIN}&path=%2Ftrojan-argo%3Fed%3D2048#{NAME}-{ISP}
    """
    
    with open(os.path.join(F_PATH, 'list.txt'), 'w', encoding='utf-8') as list_file:
        list_file.write(list_txt)

    sub_txt = base64.b64encode(list_txt.encode('utf-8')).decode('utf-8')
    with open(os.path.join(F_PATH, 'sub.txt'), 'w', encoding='utf-8') as sub_file:
        sub_file.write(sub_txt)
        
    try:
        with open(os.path.join(F_PATH, 'sub.txt'), 'rb') as file:
            sub_content = file.read()
        print(f"\n{sub_content.decode('utf-8')}")
    except FileNotFoundError:
        print(f"sub.txt not found")
    
    print(f'\n{F_PATH}/sub.txt saved successfully')
    time.sleep(45)  
 
    
    files_to_delete = ['npm', 'web', 'bot', 'boot.log', 'list.txt', 'config.json', 'tunnel.yml', 'tunnel.json']
    for file_to_delete in files_to_delete:
        F_PATH_to_delete = os.path.join(F_PATH, file_to_delete)
        if os.path.exists(F_PATH_to_delete):
            try:
                os.remove(F_PATH_to_delete)
                
            except Exception as e:
                print(f"Error deleting {F_PATH_to_delete}: {e}")
        else:
            print(f"{F_PATH_to_delete} doesn't exist, skipping deletion")

    print('\033c', end='')
    print('App is running')
    print('Thank you for using this script, enjoy!')
         

def start_server():
    download_files_and_run()
    extract_domains()
    
start_server()


has_logged_empty_message = False

def visit_project_page():
    try:
        if not P_URL or not INTERVAL_SECONDS:
            global has_logged_empty_message
            if not has_logged_empty_message:
                print("URL or TIME variable is empty, Skipping visit web")
                has_logged_empty_message = True
            return

        response = requests.get(P_URL)
        response.raise_for_status() 

        
        print("Page visited successfully")
        print('\033c', end='')
    except requests.exceptions.RequestException as error:
        print(f"Error visiting project page: {error}")

if __name__ == "__main__":
    while True:
        visit_project_page()
        time.sleep(INTERVAL_SECONDS)
