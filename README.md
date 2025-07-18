Fork 此项目，取消Copy the main branch only的勾选使其fork所有分支

在huggingface创建空白docker项目,获取域名，格式：账户名-项目名.hf.space 将域名填写到app.js里的DOMAIN变量

在app.js里填写好uuid或哪吒等变量后，全选代码到 https://obfuscator.io 混淆，选Medium或以上混淆级别，将混淆后的代码全选粘贴到app.js保存，去Actions运行构建镜像(首次需要手动，之后每次修改自动构建)

在huggingface创建空白docker项目，创建Dockerfile文件，里面填写FROM 自己的镜像地址 保存即可
