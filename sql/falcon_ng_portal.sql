 set names utf8;

drop database if exists falcon_ng_portal;
create database falcon_ng_portal;
use falcon_ng_portal;

CREATE TABLE `dept` (
  `id` int unsigned not null AUTO_INCREMENT,
  `ident` varchar(255) not null default '',
  `remark` varchar(255) not null default '',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`ident`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `dept_boss` (
  `dept_id` int unsigned not null,
  `user_id` int unsigned not null comment 'boss userid',
  KEY (`dept_id`),
  KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `node` (
  `id` int unsigned not null AUTO_INCREMENT,
  `pid` int unsigned not null,
  `name` varchar(64) not null,
  `path` varchar(255) not null,
  `leaf` int(1) not null,
  `type` int(1) not null default 0 comment '0:normal, 1:for k8s',
  PRIMARY KEY (`id`),
  KEY (`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `host` (
  `id` int unsigned not null AUTO_INCREMENT,
  `sn` char(128) not null default '',
  `ip` char(15) not null,
  `name` varchar(64) not null default '',
  `dept_id` int unsigned not null default 0 comment 'belong to dept',
  `root_remark` varchar(255) not null default '',
  `ordi_remark` varchar(255) not null default '' comment 'ordinary remark',
  `cpu` varchar(255) not null default '',
  `mem` varchar(255) not null default '',
  `disk` varchar(255) not null default '',
  `type` int(1) not null default 0 comment '0:host, 1:container',
  PRIMARY KEY (`id`),
  KEY (`ip`),
  KEY (`sn`),
  KEY (`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `node_host` (
  `node_id` int unsigned not null,
  `host_id` int unsigned not null,
  KEY(`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `node_role` (
  `id` int unsigned not null AUTO_INCREMENT,
  `node_id` int unsigned not null,
  `user_id` int unsigned not null,
  `role` int unsigned not null,
  PRIMARY KEY(`id`),
  KEY(`node_id`),
  KEY(`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

