# 前端接口

`POST /api/portal/auth/login`

校验用户登录信息的接口，is_ldap=0表示不使用LDAP账号验证，is_ldap=1表示使用LDAP账号验证

```
{
    "username": "",
    "password": "",
    "is_ldap": 0
}
```

---

`GET /api/portal/auth/logout`

退出当前账号，如果请求成功，前端需要跳转到登录页面

---

`GET /api/portal/self/profile`

获取个人信息，可以用此接口校验用户是否登录了

---

`PUT /api/portal/self/profile`

更新个人信息

```
{
    "dispname": "",
    "phone": "",
    "email": "",
    "im": ""
}
```

---

`PUT /api/portal/self/password`

更新个人密码，新密码输入两次做校验，在前端完成

```
{
    "oldpass": "",
    "newpass": ""
}
```

---

`GET /api/portal/user`

获取用户列表，支持搜索，搜索条件参数是query，每页显示条数是limit，页码是p，如果当前用户是root，则展示相关操作按钮，如果不是，则所有按钮不展示，只是查看

---

`POST /api/portal/user`

root账号新增一个用户，is_root字段表示新增的这个用户是否是个root

```
{
    "username": "",
    "password": "",
    "dispname": "",
    "phone": "",
    "email": "",
    "im": "",
    "is_root": 0
}
```

---

`GET /api/portal/user/:id/profile`

获取某个人的信息

---

`PUT /api/portal/user/:id/profile`

root账号修改某人的信息

```
{
    "dispname": "",
    "phone": "",
    "email": "",
    "im": "",
    "is_root": 0
}
```

---

`PUT /api/portal/user/:id/password`

root账号重置某人的密码，输入两次新密码保证一致的校验由前端来做

```
{
    "password": ""
}
```

---

`DELETE /api/portal/user/:id`

root账号来删除某个用户

---

`GET /api/portal/team`

获取团队列表，支持搜索，搜索条件参数是query，每页显示条数是limit，页码是p

---

`POST /api/portal/team`

创建团队，mgmt=0表示成员管理制，mgmt=1表示管理员管理制，admins是团队管理员的id列表，members是团队普通成员的id列表

```
{
    "ident": "",
    "name": "",
    "mgmt: 0,
    "admins": [],
    "members": []
}
```

---

`PUT /api/portal/team/:id`

修改团队信息

```
{
    "ident": "",
    "name": "",
    "mgmt: 0,
    "admins": [],
    "members": []
}
```

---

`DELETE /api/portal/team/:id`

删除团队

---

