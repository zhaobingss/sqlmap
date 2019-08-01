CREATE TABLE `sys_src` (
    `id` bigint(10) NOT NULL AUTO_INCREMENT COMMENT '资源ID',
    `pid` bigint(10) NOT NULL DEFAULT '0' COMMENT '父资源ID',
    `type` varchar(20) DEFAULT NULL COMMENT '资源类型',
    `name` varchar(20) DEFAULT NULL COMMENT '资源名称',
    `code` varchar(32) NOT NULL COMMENT '资源编码（必填用来标识权限）',
    `description` varchar(100) DEFAULT NULL COMMENT '资源描述',
    `url` varchar(200) DEFAULT NULL COMMENT '资源路径',
    `icon` varchar(64) DEFAULT NULL COMMENT '资源图标',
    `seq` int(4) DEFAULT NULL COMMENT '资源排序',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '资源创建时间',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '资源修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_code` (`code`) USING BTREE,
    KEY `idx_pid` (`pid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=39 DEFAULT CHARSET=utf8