# poll
A TCP link pool with a link listening rebuild capability  // 一个拥有链接监听重建能力的TCP链接池

# demo 阶段
当链接池中的链接被对端服务器正常或者异常关闭时，连接池会自动监听并进行新链接创建，防止链接池对外输出已经关闭的链接
