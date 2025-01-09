## 项目说明
#### 服务依赖
mysql    
redis  
golang@v1.13

#### 代码拉取
1.git clone https://github.com/sdgdg556/airbnb_data_collector.git    
2.cd data_collector    
3.go mod tidy   
4.go mod vendor    

#### 服务运行
1.用docker或者本地启动启动mysql和redis实例   
2.将sql目录下的建表语句在mysql实例上跑一下   
3.将本地redis和mysql配置更新道config目录下的config.yaml文件中   
4.执行go build -o airbnb-cli    
5.执行命令./airbnb-cli start [consumer|producer] [--workers=N] --queue=<your-queue-server> --data tasks.json