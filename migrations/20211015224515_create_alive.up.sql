CREATE TABLE  IF NOT EXISTS `alive` (
 `alive_ip` int(10) unsigned NOT NULL,
 `alive_is` tinyint(1) NOT NULL,
 `alive_ts` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
 `alive_mac` char(20) NOT NULL DEFAULT '0',
 PRIMARY KEY (`alive_ip`) USING BTREE,
 KEY `alive_ts` (`alive_ts`) USING BTREE,
 KEY `alive_is_ip` (`alive_is`,`alive_ip`) USING BTREE
) ENGINE=MEMORY DEFAULT CHARSET=utf8