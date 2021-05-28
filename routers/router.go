package routers

import (
	"testbeego/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/", &controllers.MainController{}, "get:ShowGet;post:Post")

	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:HandlePost")

	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	//文章列表页访问
	beego.Router("/showArticleList", &controllers.ArticleController{}, "get:ShowArticleList")
	//添加文章
	beego.Router("/addArticle", &controllers.ArticleController{}, "get:ShowAddArticle;post:HandleAddArticle")
	//显示文章详情
	beego.Router("/showArticleDetail", &controllers.ArticleController{}, "get:ShowArticleDetail")

}
