 set names utf8;

drop database if exists falcon_ng_portal;
create database falcon_ng_portal;
use falcon_ng_portal;

CREATE TABLE `node` (
  `id` int unsigned not null AUTO_INCREMENT,
  `pid` int unsigned not null,
  `name` varchar(64) not null,
  `path` varchar(255) not null,
  `leaf` int(1) not null,
  `type` int(1) not null default 0 comment '0:normal, 1: thirdparty',
  `note` varchar(128) not null default '',
  PRIMARY KEY (`id`),
  KEY (`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `endpoint` (
  `id` int unsigned not null AUTO_INCREMENT,
  `ident` varchar(255) not null,
  `alias` varchar(255) not null default '',
  PRIMARY KEY (`id`),
  KEY (`ident`),
  KEY (`alias`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `node_endpoint` (
  `node_id` int unsigned not null,
  `endpoint_id` int unsigned not null,
  KEY(`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
