-- 创建聊天机器人表
CREATE TABLE IF NOT EXISTS `chatbot` (
    `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '聊天机器人ID',
    `name` varchar(100) NOT NULL COMMENT '机器人名称',
    `personality` text COMMENT '性格设定',
    `background` text COMMENT '背景设定',
    `system_prompt` text COMMENT '系统提示词',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_created_at` (`created_at`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='聊天机器人表';

-- 创建对话历史表
CREATE TABLE IF NOT EXISTS `chat_history` (
    `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '对话历史ID',
    `chatbot_id` int(11) NOT NULL COMMENT '聊天机器人ID',
    `role` varchar(20) NOT NULL COMMENT '角色：user 或 assistant',
    `content` text NOT NULL COMMENT '消息内容',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_chatbot_id` (`chatbot_id`) USING BTREE,
    KEY `idx_created_at` (`created_at`) USING BTREE,
    CONSTRAINT `fk_chat_history_chatbot` FOREIGN KEY (`chatbot_id`) REFERENCES `chatbot` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='对话历史表';

