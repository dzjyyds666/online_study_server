drop table if exists user;
create table user
(
    user_id         varchar(16) primary key not null comment '用户id',
    nick_name       varchar(16)             not null comment '昵称',
    avatar          varchar(64)             not null comment '头像',
    gender          varchar(1)              not null comment '性别',
    email           varchar(32) unique      not null comment '邮箱',
    password        varchar(64)             not null comment '密码',
    create_ts       int                     not null comment '创建时间',
    role            varchar(1)              not null comment '角色',
    status          varchar(1)              not null comment '状态',
    last_login_time int                     not null comment '最后登录时间',
    last_login_ip   varchar(32)             not null comment '最后登录ip',
)

