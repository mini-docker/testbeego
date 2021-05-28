package controllers

import (
	"fmt"
	"math"
	"path"
	"testbeego/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	beego "github.com/beego/beego/v2/server/web"
)

type ArticleController struct {
	beego.Controller
}

//展示文章列表页
func (this *ArticleController) ShowArticleList() {
	//获取数据
	//高级查询
	//指定表
	o := orm.NewOrm()
	qs := o.QueryTable("Article") //queryseter
	var articles []models.Article
	//_,err := qs.All(&articles)
	//if err != nil{
	//	fmt.Printf("查询数据错误")
	//}
	//查询总记录数
	count, _ := qs.Count()
	//获取总页数
	pageSize := 2

	pageCount := math.Ceil(float64(count) / float64(pageSize))
	//获取页码
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}

	//获取数据
	//作用就是获取数据库部分数据,第一个参数，获取几条,第二个参数，从那条数据开始获取,返回值还是querySeter
	//起始位置计算
	start := (pageIndex - 1) * pageSize

	qs.Limit(pageSize, start).All(&articles)

	//传递数据
	this.Data["pageIndex"] = pageIndex
	this.Data["pageCount"] = int(pageCount)
	this.Data["count"] = count
	this.Data["articles"] = articles
	this.TplName = "index.html"
}

//展示添加文章页面
func (this *ArticleController) ShowAddArticle() {
	this.TplName = "add.html"
}

//获取添加文章数据
func (this *ArticleController) HandleAddArticle() {
	//1.获取数据
	articleName := this.GetString("articleName")
	content := this.GetString("content")

	//2校验数据
	if articleName == "" || content == "" {
		this.Data["errmsg"] = "添加数据不完整"
		this.TplName = "add.html"
		return
	}

	//处理文件上传
	file, head, err := this.GetFile("uploadname")
	defer file.Close()
	if err != nil {
		this.Data["errmsg"] = "文件上传失败"
		this.TplName = "add.html"
		return
	}

	//1.文件大小
	if head.Size > 5000000 {
		this.Data["errmsg"] = "文件太大，请重新上传"
		this.TplName = "add.html"
		return
	}

	//2.文件格式
	//a.jpg
	ext := path.Ext(head.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		this.Data["errmsg"] = "文件格式错误。请重新上传"
		this.TplName = "add.html"
		return
	}

	//3.防止重名
	fileName := time.Now().Format("2006-01-02-15:04:05") + ext
	//存储
	this.SaveToFile("uploadname", "./static/img/"+fileName)

	//3.处理数据
	//插入操作
	o := orm.NewOrm()

	var article models.Article
	article.ArtiName = articleName
	article.Acontent = content
	article.Aimg = "/static/img/" + fileName

	o.Insert(&article)

	//4.返回页面
	this.Redirect("/showArticleList", 302)
}

//展示文章详情页面
func (this *ArticleController) ShowArticleDetail() {
	//获取数据
	id, er := this.GetInt("articleId")
	//数据校验
	if er != nil {
		fmt.Printf("传递的链接错误")
	}
	//操作数据
	o := orm.NewOrm()
	var article models.Article
	article.Id = id

	o.Read(&article)

	//修改阅读量
	article.Acount += 1
	o.Update(&article)

	//返回视图页面
	this.Data["article"] = article
	this.TplName = "content.html"
}
