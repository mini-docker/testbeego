package controllers

import (
	"fmt"
	"testbeego/models"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
)

type UserController struct {
	beego.Controller
}

//显示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

//处理注册数据
func (this *UserController) HandlePost() {
	//1.获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")

	//fmt.Printf(userName,pwd)
	//2.校验数据
	if userName == "" || pwd == "" {
		this.Data["errmsg"] = "注册数据不完整，请重新注册"
		fmt.Printf("注册数据不完整，请重新注册")
		this.TplName = "register.html"
		return
	}
	//3.操作数据
	//获取ORM对象
	o := orm.NewOrm()
	//获取插入对象
	var user models.User
	//给插入对象赋值
	user.Name = userName
	user.PassWord = pwd
	//插入
	o.Insert(&user)
	//返回结果

	//4.返回页面
	//this.Ctx.WriteString("注册成功")
	this.Redirect("/login", 302)
	//this.TplName = "login.html"
}

//展示登录页面
func (this *UserController) ShowLogin() {
	this.Data["data"] = "heheh"
	this.TplName = "login.html"
}
func (this *UserController) HandleLogin() {
	//1.获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")
	//2.校验数据
	if userName == "" || pwd == "" {
		this.Data["errmsg"] = "登录数据不完整"
		this.TplName = "login.html"
		return
	}

	//3.操作数据
	//1。获取ORM对象
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		this.Data["errmsg"] = "用户不存在"
		this.TplName = "login.html"
		return
	}
	if user.PassWord != pwd {
		this.Data["errmsg"] = "密码错误"
		this.TplName = "login.html"
		return
	}

	//4.返回页面
	//this.Ctx.WriteString("登录成功")
	this.Redirect("/showArticleList", 302)
}
