package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

//spider bilibili

type JsonModel struct {
	//Ttl  int  `json:"ttl"`
	Data Data `json:"data"`
}

type Data struct {
	List List `json:"list"`
	Page Page `json:"page"`
}

type List struct {
	Vlist []Vlist `json:"vlist"`
}

type Vlist struct {
	Author       string `json:"author"`
	Comment      int32  `json:"comment"` //评论数
	Play         int32  `json:"play"`    //播放量
	Pic          string `json:"pic"`
	Title        string `json:"title"`
	Created      int64  `json:"created"`
	Video_review int    `json:"video_review"` //弹幕数
	Bvid         string `json:"bvid"`         //视频ID
}

type Page struct {
	Pn    int `json:"pn"`
	Ps    int `json:"ps"`
	Count int `json:"count"`
}

//1885078 --Nya酱
//163637592 --何同学
//1951371956 --Juli刘
var upId int = 1951371956

func main() {
	indexUrl := fmt.Sprintf("https://api.bilibili.com/x/space/arc/search?mid=%d&ps=30&tid=0&pn=1&keyword=&order=pubdate&jsonp=jsonp", upId)
	maxPageIndex := GetMaxPageIndex(indexUrl)
	SpiderUrl(maxPageIndex)
}

func GetMaxPageIndex(url string) int {
	jsonModel, _ := GetUrlBody(url)
	pageCount := jsonModel.Data.Page.Count
	pageSize := jsonModel.Data.Page.Ps
	if pageSize == 0 {
		panic("pageSize is zero")
	}

	return int(math.Ceil(float64(pageCount) / float64(pageSize)))
}

func GetUrlBody(url string) (JsonModel, bool) {
	var jsonModel JsonModel
	isSuccess := true
	res, erro := http.Get(url)
	if erro != nil {
		log.Printf("Request erro:%s", url)
	} else {
		by, _ := ioutil.ReadAll(res.Body)
		erro1 := json.Unmarshal([]byte(by), &jsonModel)
		if erro1 != nil {
			fmt.Println("Unmarshal erro")
			isSuccess = false
		}
	}
	defer func() {
		if erro := recover(); erro != nil {
			log.Println("Get Body erro")
			isSuccess = false
		}
	}()
	return jsonModel, isSuccess
}

func SpiderUrl(maxPageIndex int) {
	if maxPageIndex != 0 {
		for index := 1; index <= maxPageIndex; index++ {
			spiderUrl := fmt.Sprintf("https://api.bilibili.com/x/space/arc/search?mid=%d&ps=30&tid=0&pn=%d&keyword=&order=pubdate&jsonp=jsonp", upId, index)
			rand.Seed(time.Now().Unix())
			sleepRandom := rand.Intn(30)
			time.Sleep(time.Duration(sleepRandom) * time.Second)
			log.Printf("sleep %d second:%s", sleepRandom, spiderUrl)
			JsonModel, isSuccess := GetUrlBody(spiderUrl)
			if isSuccess {
				InsertToDB(spiderUrl, &JsonModel)
			} else {
				log.Printf("Get JsonModel Erro:%s", spiderUrl)
			}

		}
	} else {
		log.Println("maxPageIndex is zero")
	}
}

func InsertToDB(url string, spiderModel *JsonModel) {
	var server = "192.168.83.144"
	var port = 1433
	var user = "sa"
	var password = "kingdee@2018"
	var database = "Test"
	sqlStr := GetSql(url, spiderModel)
	//fmt.Println(sqlStr)
	//连接字符串
	connString := fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s", server, port, database, user, password)
	// if isdebug {
	// 	fmt.Println(connString)
	// }
	//建立连接
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open Connection failed:", err.Error())
	}
	defer conn.Close()

	//产生查询语句的Statement
	result, err := conn.Exec(sqlStr)
	if err != nil {
		log.Fatal("insert failed:", err.Error())
	}
	effectRow, err := result.RowsAffected()
	if err != nil {
		log.Fatal("insert failed:", err.Error())
	}
	if effectRow > 0 {
		log.Printf("成功抓取：%s", url)
	}

}

func GetSql(url string, spiderModel *JsonModel) string {
	var sqlStr string

	spiderInfos := spiderModel.Data.List.Vlist
	for _, v := range spiderInfos {
		createTime := time.Unix(v.Created, 0).Format("2006-01-02 15:04:05")
		sqlStr += fmt.Sprintf(`
		INSERT INTO [dbo].[T_UPInfo]
					([FFromUrl]
					,[FAuthor]
					,[FPic]
					,[FTitle]
					,[FComment]
					,[FPalyCount]
					,[FCreated]
					,[FVideo_review]
					,[FBvid]
					,[FAddTime])
		VALUES	('%s','%s','%s','%s','%d','%d','%s','%d','%s',GETDATE())
		`, url, v.Author, v.Pic, v.Title, v.Comment, v.Play, createTime, v.Video_review, v.Bvid)
	}

	return sqlStr
}

func getJson() string {
	return `{"code":0,"message":"0","ttl":1,"data":{"list":{"tlist":{"4":{"tid":4,"count":13,"name":"游戏"},"129":{"tid":129,"count":2,"name":"舞蹈"},"155":{"tid":155,"count":101,"name":"时尚"},"160":{"tid":160,"count":66,"name":"生活"},"211":{"tid":211,"count":61,"name":"美食"}},"vlist":[{"comment":7517,"typeid":21,"play":2534819,"pic":"http://i1.hdslb.com/bfs/archive/956ca4e726c4b6fd05f0cff1c34a1b3506e91710.jpg","subtitle":"","description":"","copyright":"1","title":"传说中的天津味女仆咖啡店被我找到了!没想到服务巨好!","review":0,"author":"nya酱的一生","mid":1885078,"created":1620722374,"length":"10:59","video_review":32451,"aid":375572496,"bvid":"BV1uo4y1m7CA","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":24711,"typeid":21,"play":2854078,"pic":"http://i1.hdslb.com/bfs/archive/4b2073c00c8240c7bcfa18eb32b6b87754d45e69.jpg","subtitle":"","description":"-","copyright":"1","title":"瘫痪在床几个月后，我决定一辈子不生孩子了。","review":0,"author":"nya酱的一生","mid":1885078,"created":1619854937,"length":"12:24","video_review":42898,"aid":460344726,"bvid":"BV1Y5411g7Y8","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":2337,"typeid":157,"play":786719,"pic":"http://i0.hdslb.com/bfs/archive/01519d61605aeb721f68f921b16a9b268d16a6cd.jpg","subtitle":"","description":"","copyright":"1","title":"口罩都遮不住这些口红的美？！粉丝推荐口红系列！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1608197919,"length":"12:17","video_review":10879,"aid":500721722,"bvid":"BV1CK411u71f","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":3754,"typeid":21,"play":1268218,"pic":"http://i2.hdslb.com/bfs/archive/774c64c5d958b2855a380140dd8bff74f5766020.jpg","subtitle":"","description":"-","copyright":"1","title":"让男友看我写给前任的青春疼痛QQ文学！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1607068704,"length":"14:01","video_review":13760,"aid":373030467,"bvid":"BV1yZ4y1G7Hg","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":2450,"typeid":157,"play":1113476,"pic":"http://i1.hdslb.com/bfs/archive/2ea9d502b93f3b4b03827e3f86dd1abb880e529f.jpg","subtitle":"","description":"","copyright":"1","title":"滴在胸罩里能让ru晕美白的产品？！近期垃圾大合集！","review":0,"author":"nya酱的一生","mid":1885078,"created":1606123810,"length":"09:12","video_review":10232,"aid":457771456,"bvid":"BV165411V7Au","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":1410,"typeid":213,"play":854982,"pic":"http://i1.hdslb.com/bfs/archive/b40583d3835f834fd1dbcc7077d71a1e1026bcab.jpg","subtitle":"","description":"吃完我真的很绝望。。。。。","copyright":"1","title":"世界顶级餐厅价值近2000的外卖长啥样？吃完我只想口吐芬芳。。。。。","review":0,"author":"nya酱的一生","mid":1885078,"created":1605077590,"length":"09:49","video_review":6170,"aid":800180204,"bvid":"BV1Uy4y1B7Hc","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":3292,"typeid":164,"play":1165865,"pic":"http://i0.hdslb.com/bfs/archive/6d79c5a35f748a21fd2f950ee93de3fd7e3e50be.jpg","subtitle":"","description":"钙尔奇为合作内容～ヾ(^▽^*)))","copyright":"1","title":"126斤的我决定再也不减肥了【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1604394041,"length":"10:48","video_review":6797,"aid":670197332,"bvid":"BV1qa4y1s7Wf","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":4931,"typeid":157,"play":653236,"pic":"http://i0.hdslb.com/bfs/archive/97e563d79fc9a253262e60ee98790f8c91a5bcb1.jpg","subtitle":"","description":"该视频为合作视频~\n所有产品都是细心使用后挑选出来的~希望你们也喜欢~","copyright":"1","title":"这些好东西错过真的要等一年！nya的究极薅法来了！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1604030629,"length":"13:04","video_review":3990,"aid":457512911,"bvid":"BV1w5411L7BQ","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":1715,"typeid":138,"play":628612,"pic":"http://i2.hdslb.com/bfs/archive/c89666927e415ab21240335139de7a9e2bfa0b0c.jpg","subtitle":"","description":"-","copyright":"1","title":"让你们见识下沉浸式万圣节，不一样的你比划我猜！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1603972696,"length":"09:30","video_review":11512,"aid":457561410,"bvid":"BV1T5411L7WU","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":2600,"typeid":21,"play":1234705,"pic":"http://i0.hdslb.com/bfs/archive/8722156e8bc43d86e9f8a974b3487af7b6de051b.jpg","subtitle":"","description":"","copyright":"1","title":"为什么种菜成了我的噩梦，参观下我惨不忍睹的菜园。。。。","review":0,"author":"nya酱的一生","mid":1885078,"created":1603711731,"length":"09:53","video_review":8223,"aid":287516088,"bvid":"BV1of4y1B7Go","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":3649,"typeid":21,"play":745128,"pic":"http://i1.hdslb.com/bfs/archive/3f5c6b50eb589f4ddaab50d13b720a6cc156705f.jpg","subtitle":"","description":"","copyright":"1","title":"吓尿!被困海上的我第一次有了靠近死亡的体验！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1603196191,"length":"12:19","video_review":8936,"aid":797611023,"bvid":"BV1dy4y1r7vZ","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":7819,"typeid":21,"play":2070257,"pic":"http://i1.hdslb.com/bfs/archive/5c9a759443eed6aaa621aee010661cd7f3934549.jpg","subtitle":"","description":"嘿嘿","copyright":"1","title":"竟然有营销号敢黑我纹身？！nya要开怼了！","review":0,"author":"nya酱的一生","mid":1885078,"created":1596872649,"length":"11:26","video_review":37329,"aid":884146506,"bvid":"BV1kK4y1v7vu","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":10037,"typeid":157,"play":1731460,"pic":"http://i0.hdslb.com/bfs/archive/23ff65eefda1e6fdd0aa521d8058510fdc667d6c.jpg","subtitle":"","description":"neinei的色号是nw01哦~❤️\nBlank ME小银盒气垫为合作内容?","copyright":"1","title":"伪装成绿茶，男人会不会为我疯狂? 伪装成绿茶过一天！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1596107949,"length":"10:24","video_review":39641,"aid":839083438,"bvid":"BV1Z54y1S7ve","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":5826,"typeid":157,"play":1751665,"pic":"http://i1.hdslb.com/bfs/archive/27002eadb4b0d2833f3560abbd92f4832d269665.jpg","subtitle":"","description":"","copyright":"1","title":"挑战一天时间装成三种不同人设网红！nya酱一人分饰三角累惨了!","review":0,"author":"nya酱的一生","mid":1885078,"created":1595062816,"length":"10:49","video_review":50567,"aid":371467297,"bvid":"BV1iZ4y1T7aj","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":1715,"typeid":157,"play":826126,"pic":"http://i0.hdslb.com/bfs/archive/653543587aca4d55d2ba9a9d5abbd596ddeffdf2.jpg","subtitle":"","description":"","copyright":"1","title":"用闻所未闻的非洲护肤品给自己做Spa！有些产品竟然出奇好用！【nya酱】\u001c","review":0,"author":"nya酱的一生","mid":1885078,"created":1591944375,"length":"07:55","video_review":12831,"aid":753499915,"bvid":"BV1uk4y1z7wZ","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":1824,"typeid":157,"play":687052,"pic":"http://i1.hdslb.com/bfs/archive/8f035ef7549f8f1d10168f821e93ac4645856d30.jpg","subtitle":"","description":"美即面膜 pf元气小丸子口红为合作内容","copyright":"1","title":"耳朵专用的蜡烛你试过吗？这次的购物分享有点野啊","review":0,"author":"nya酱的一生","mid":1885078,"created":1590408042,"length":"11:46","video_review":9466,"aid":968277057,"bvid":"BV1Kp4y1X769","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":4006,"typeid":21,"play":958514,"pic":"http://i1.hdslb.com/bfs/archive/a07a5a5e3f8a2babd4786bbd96c23a534e562723.jpg","subtitle":"","description":"我佛了","copyright":"1","title":"挑战只用“恶魔之语”温州话和天津人沟通一天！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1590307212,"length":"10:18","video_review":16796,"aid":925846267,"bvid":"BV1GT4y1g77d","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":4726,"typeid":21,"play":705302,"pic":"http://i2.hdslb.com/bfs/archive/7c661b7b1057e4e5d8b4fb608f342a65347a9f5b.jpg","subtitle":"","description":"雅顿橘灿系列为合作内容","copyright":"1","title":"如果穿越回高中，我的一天是怎么样的？【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1589194616,"length":"11:21","video_review":8590,"aid":925693444,"bvid":"BV1xT4y1u7iR","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":7818,"typeid":158,"play":3075518,"pic":"http://i2.hdslb.com/bfs/archive/00e2bf7380d726e508ab6a67054c6189de8e17bc.jpg","subtitle":"","description":"","copyright":"1","title":"普通身材亚洲女孩尝试性感辣妹风，太辣了！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1588843224,"length":"14:11","video_review":146781,"aid":285590597,"bvid":"BV1Cf4y1m7TF","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":4242,"typeid":21,"play":1002689,"pic":"http://i2.hdslb.com/bfs/archive/6d8f1179bdafe421d42ed82fcf39a841eaf1e104.jpg","subtitle":"","description":"对fresh鸳鸯锅面膜感兴趣的小朋友可以去关注一下在小黑盒首发哦！\nfresh鸳鸯锅面膜为合作内容","copyright":"1","title":"零基础三天学yes ok能学成啥样！练到崩溃！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1587816053,"length":"12:28","video_review":17993,"aid":795442918,"bvid":"BV1MC4y1W7e3","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":2449,"typeid":76,"play":859916,"pic":"http://i1.hdslb.com/bfs/archive/dbe4a6a667a1f36789300a1f7187b83de48f3461.jpg","subtitle":"","description":"嗝","copyright":"1","title":"这盘注入灵魂的至尊芝士咖喱有点过于美味了！！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1587734269,"length":"11:29","video_review":4920,"aid":285470129,"bvid":"BV11f4y1S79w","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":6228,"typeid":21,"play":1517380,"pic":"http://i0.hdslb.com/bfs/archive/aadcf1d50c0b4e519e4b58f608b19d927a02518d.jpg","subtitle":"","description":"薇婷为合作内容","copyright":"1","title":"让你们见识下皇家级别的仪式感约会【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1586775366,"length":"13:14","video_review":30926,"aid":242733404,"bvid":"BV1Le41147qi","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":2341,"typeid":17,"play":654937,"pic":"http://i1.hdslb.com/bfs/archive/f3847f96ca1da734766df850b0cb0aff852a069c.jpg","subtitle":"","description":"超级鸡马","copyright":"1","title":"我不开玩笑，这互搞游戏情侣千万不要玩【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1586090039,"length":"29:43","video_review":8970,"aid":667744314,"bvid":"BV1ta4y1x79H","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":9117,"typeid":21,"play":996066,"pic":"http://i0.hdslb.com/bfs/archive/201a6ea5b5c846a7be2eceb64341b17db0a883bf.jpg","subtitle":"","description":"","copyright":"1","title":"没想到我会离癌症这么近，是时候聊聊我的乳腺结节手术了。","review":0,"author":"nya酱的一生","mid":1885078,"created":1585033206,"length":"14:15","video_review":34444,"aid":98559307,"bvid":"BV1tE411c7B4","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":3913,"typeid":157,"play":980410,"pic":"http://i2.hdslb.com/bfs/archive/7970ff8046f40161f3f1e54adba802510c506480.jpg","subtitle":"","description":"欧莱雅紫熨斗为合作内容","copyright":"1","title":"20岁到30岁这十年，时间改变了我什么？晨间grwm！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1583748007,"length":"15:24","video_review":24715,"aid":94797124,"bvid":"BV1wE411u7vt","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":2898,"typeid":157,"play":728435,"pic":"http://i0.hdslb.com/bfs/archive/2425662d93f585a7e6c23f6b01fc5e07ab91045f.jpg","subtitle":"","description":"-","copyright":"1","title":"如果直男和美妆博主比拼直播卖货！究竟谁更带货！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1583409603,"length":"16:55","video_review":13421,"aid":93773441,"bvid":"BV17E411x7PK","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":4014,"typeid":21,"play":1154804,"pic":"http://i2.hdslb.com/bfs/archive/1d05675fbb3c319a0272e5c145e0a62bd05a3baa.jpg","subtitle":"","description":"","copyright":"1","title":"因为疑似新冠肺炎，我被澳洲医院隔离了【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1582726775,"length":"11:14","video_review":6994,"aid":91955082,"bvid":"BV1o7411K7oa","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":3189,"typeid":192,"play":672061,"pic":"http://i0.hdslb.com/bfs/archive/8da3499674489cfa92c94c2d3c173a9d870aba52.jpg","subtitle":"","description":"难以置信","copyright":"1","title":"面目全非！我上了石原里美同款时尚杂志？！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1581768682,"length":"11:56","video_review":13801,"aid":89227728,"bvid":"BV1X7411J7h3","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":3832,"typeid":157,"play":878949,"pic":"http://i1.hdslb.com/bfs/archive/8848a122db57387c3482d9c29d098fcf3176e024.jpg","subtitle":"","description":"","copyright":"1","title":"难度爆表！直男根据女朋友的提示买化妆品做情人节礼物！【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1581580951,"length":"14:33","video_review":18899,"aid":88647272,"bvid":"BV1h741137CU","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0},{"comment":2049,"typeid":21,"play":513361,"pic":"http://i0.hdslb.com/bfs/archive/ae39434f0dbbebd51627defcb33b1e2b90b2f8e5.jpg","subtitle":"","description":"喵喵喵？\n喵？\n喵！！！！！！\nwhhszwcnm：）","copyright":"1","title":"北海道的企鹅被迫游行！雪国旭川动物园vlog～【nya酱】","review":0,"author":"nya酱的一生","mid":1885078,"created":1580477718,"length":"14:13","video_review":9140,"aid":85962574,"bvid":"BV1A7411z7Vo","hide_click":false,"is_pay":0,"is_union_video":0,"is_steins_gate":0,"is_live_playback":0}]},"page":{"pn":1,"ps":30,"count":243},"episodic_button":{"text":"播放全部","uri":"//www.bilibili.com/medialist/play/1885078?from=space"}}}`
}
