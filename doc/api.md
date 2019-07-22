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