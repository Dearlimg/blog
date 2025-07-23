-- 1. 删除评论表（依赖 user、video）
DROP TABLE IF EXISTS `comment`;

-- 2. 删除点赞表（依赖 user、video）
DROP TABLE IF EXISTS `like`;

-- 3. 删除消息表（依赖 user）
DROP TABLE IF EXISTS `message`;

-- 4. 删除关系表（依赖 user）
DROP TABLE IF EXISTS `relation`;

-- 5. 删除视频表（依赖 user）
DROP TABLE IF EXISTS `video`;

-- 6. 删除用户表（无依赖，最基础）
DROP TABLE IF EXISTS `user`;