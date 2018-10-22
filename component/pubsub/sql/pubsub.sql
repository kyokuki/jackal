
-- ----------------------------
-- Table structure for pubsub_affiliations
-- ----------------------------
DROP TABLE IF EXISTS `pubsub_affiliations`;
CREATE TABLE `pubsub_affiliations` (
  `node_id` bigint(20) NOT NULL,
  `jid_id` bigint(20) NOT NULL,
  `affiliation` varchar(20) NOT NULL,
  PRIMARY KEY (`node_id`,`jid_id`),
  UNIQUE KEY `node_id_2` (`node_id`,`jid_id`) USING HASH,
  KEY `node_id` (`node_id`) USING HASH,
  KEY `jid_id` (`jid_id`) USING HASH

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for pubsub_items
-- ----------------------------
DROP TABLE IF EXISTS `pubsub_items`;
CREATE TABLE `pubsub_items` (
  `node_id` bigint(20) NOT NULL,
  `id` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
  `id_sha1` char(40) NOT NULL,
  `creation_date` datetime DEFAULT NULL,
  `publisher_id` bigint(20) DEFAULT NULL,
  `update_date` datetime DEFAULT NULL,
  `data` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
  PRIMARY KEY (`node_id`,`id_sha1`),
  KEY `node_id_2` (`node_id`) USING HASH,
  KEY `publisher_id` (`publisher_id`),
  KEY `node_id_id` (`node_id`,`id`(190)) USING HASH

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for pubsub_jids
-- ----------------------------
DROP TABLE IF EXISTS `pubsub_jids`;
CREATE TABLE `pubsub_jids` (
  `jid_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `jid` varchar(2049) NOT NULL,
  `jid_sha1` char(40) NOT NULL,
  PRIMARY KEY (`jid_id`),
  UNIQUE KEY `jid_sha1` (`jid_sha1`) USING HASH,
  KEY `jid` (`jid`(255)) USING HASH
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for pubsub_nodes
-- ----------------------------
DROP TABLE IF EXISTS `pubsub_nodes`;
CREATE TABLE `pubsub_nodes` (
  `node_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `service_id` bigint(20) NOT NULL,
  `name` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `name_sha1` char(40) NOT NULL,
  `type` int(11) NOT NULL,
  `title` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
  `description` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
  `creator_id` bigint(20) DEFAULT NULL,
  `creation_date` datetime DEFAULT NULL,
  `configuration` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
  `collection_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`node_id`),
  UNIQUE KEY `service_id_3` (`service_id`,`name_sha1`) USING HASH,
  KEY `service_id` (`service_id`) USING HASH,
  KEY `collection_id` (`collection_id`) USING HASH,
  KEY `creator_id` (`creator_id`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for pubsub_service_jids
-- ----------------------------
DROP TABLE IF EXISTS `pubsub_service_jids`;
CREATE TABLE `pubsub_service_jids` (
  `service_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `service_jid` varchar(2049) NOT NULL,
  `service_jid_sha1` char(40) NOT NULL,
  PRIMARY KEY (`service_id`),
  UNIQUE KEY `service_jid_sha1` (`service_jid_sha1`) USING HASH,
  KEY `service_jid` (`service_jid`(255)) USING HASH
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for pubsub_subscriptions
-- ----------------------------
DROP TABLE IF EXISTS `pubsub_subscriptions`;
CREATE TABLE `pubsub_subscriptions` (
  `node_id` bigint(20) NOT NULL,
  `jid_id` bigint(20) NOT NULL,
  `subscription` varchar(20) NOT NULL,
  `subscription_id` varchar(40) NOT NULL,
  PRIMARY KEY (`node_id`,`jid_id`),
  UNIQUE KEY `node_id_2` (`node_id`,`jid_id`) USING HASH,
  KEY `node_id` (`node_id`) USING HASH,
  KEY `jid_id` (`jid_id`) USING HASH
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;
