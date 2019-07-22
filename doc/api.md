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

`GET /api/portal/endpoint`

获取endpoint列表，用于【服务树】-【对象列表】页面，该页展示endpoint列表，搜索条件参数是query，每页显示条数是limit，页码是p，如果要做批量筛选，则同时要指定用哪个字段来筛选，只支持ident和alias，批量筛选的内容是batch

---

`POST /api/portal/endpoint`

导入endpoint，要求传入列表，每一条是ident::alias拼接在一起

```
{
    "endpoints": []
}
```

---

`PUT /api/portal/endpoint/:id`

修改一个endpoint的alias信息

```
{
    "alias": ""
}
```

---

`DELETE /api/portal/endpoint`

删除多个endpoint，用QueryString里的参数ids指定，逗号分隔

---

`GET /api/portal/endpoints/bindings`

查询endpoint的绑定关系，QueryString：idents，逗号分隔多个

---

`GET /api/portal/endpoints/bynodeids`

根据节点id查询挂载了哪些endpoint，QueryString：ids，逗号分隔的多个节点id

---

`GET /api/portal/tree`

查询整颗服务树

---

`GET /api/portal/tree/search`

根据节点路径(path)查询服务树子树

---

`POST /api/portal/node`

创建服务树节点，pid表示父节点id，leaf=0表示非叶子节点，leaf=1表示叶子节点，note是备注信息

```
{
    "pid": 0,
    "name": "",
    "leaf": 0,
    "note": ""
}
```

---

`PUT /api/portal/node/:id/name`

服务树节点改名

```
{
    "name": ""
}
```

---

`DELETE /api/portal/node/:id`

删除服务树节点

---
