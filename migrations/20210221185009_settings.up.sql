create table `settings` (
    `key` varchar(255) not null,
    `value` text not null,
    primary key (`key`)
);

insert into `settings` (`key`, `value`) values ('username', 'admin');
insert into `settings` (`key`, `value`) values ('password', '$2a$10$LpbHiC5IKXDKIwi33gmj9uipd33nMsLov0rIL9kCFw45zhf72fHme');
