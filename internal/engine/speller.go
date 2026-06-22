package engine

// Speller defines the interface for converting input codes to pinyin syllables.
type Speller interface {
	Name() string
	ToPinyin(code string) []string
}

// xiaoheLookup maps 2-character Xiaohe codes to their pinyin syllable(s).
var xiaoheLookup = map[string][]string{
	// a - zero-initial
	"aa": {"a"},
	"ah": {"ang"},
	// b - ba series
	"ba": {"ba"}, "bd": {"bai"}, "bj": {"ban"}, "bh": {"bang"}, "bc": {"bao"},
	"bw": {"bei"}, "bf": {"ben"}, "bg": {"beng"},
	"bi": {"bi"}, "bm": {"bian"}, "bn": {"biao"}, "bp": {"bie"}, "bb": {"bin"}, "bk": {"bing"},
	"bo": {"bo"}, "bu": {"bu"},
	// c - ca series
	"ca": {"ca"}, "cd": {"cai"}, "cj": {"can"}, "ch": {"cang"}, "cc": {"cao"},
	"ce": {"ce"}, "cf": {"cen"}, "cg": {"ceng"}, "ci": {"ci"},
	"cs": {"cong"}, "cz": {"cou"},
	"cu": {"cu"}, "cr": {"cuan"}, "cv": {"cui"}, "cy": {"cun"}, "co": {"cuo"},
	// d - da series
	"da": {"da"}, "dd": {"dai"}, "dj": {"dan"}, "dh": {"dang"}, "dc": {"dao"},
	"de": {"de"}, "dw": {"dei"}, "dg": {"deng"},
	"di": {"di"}, "dx": {"dia"}, "dm": {"dian"}, "dn": {"diao"}, "dp": {"die"},
	"dk": {"ding"}, "dq": {"diu"},
	"ds": {"dong"}, "dz": {"dou"},
	"du": {"du"}, "dr": {"duan"}, "dv": {"dui"}, "dy": {"dun"}, "do": {"duo"},
	// e - zero-initial
	"ee": {"e"}, "ew": {"ei"}, "eg": {"eng"}, "er": {"er"},
	// f - fa series
	"fa": {"fa"}, "fj": {"fan"}, "fh": {"fang"},
	"fw": {"fei"}, "ff": {"fen"}, "fg": {"feng"},
	"fo": {"fo"}, "fz": {"fou"}, "fu": {"fu"},
	// g - ga series
	"ga": {"ga"}, "gd": {"gai"}, "gj": {"gan"}, "gh": {"gang"}, "gc": {"gao"},
	"ge": {"ge"}, "gw": {"gei"}, "gf": {"gen"}, "gg": {"geng"},
	"gs": {"gong"}, "gz": {"gou"},
	"gu": {"gu"}, "gx": {"gua"}, "gk": {"guai"}, "gr": {"guan"}, "gl": {"guang"}, "gv": {"gui"}, "gy": {"gun"}, "go": {"guo"},
	// h - ha series
	"ha": {"ha"}, "hd": {"hai"}, "hj": {"han"}, "hh": {"hang"}, "hc": {"hao"},
	"he": {"he"}, "hw": {"hei"}, "hf": {"hen"}, "hg": {"heng"},
	"hs": {"hong"}, "hz": {"hou"},
	"hu": {"hu"}, "hx": {"hua"}, "hk": {"huai"}, "hr": {"huan"}, "hl": {"huang"}, "hv": {"hui"}, "hy": {"hun"}, "ho": {"huo"},
	// i (ch)
	"ia": {"cha"}, "id": {"chai"}, "ij": {"chan"}, "ih": {"chang"}, "ic": {"chao"},
	"ie": {"che"}, "if": {"chen"}, "ig": {"cheng"},
	"ii": {"chi"}, "is": {"chong"}, "iz": {"chou"},
	"iu": {"chu"}, "ix": {"chua"}, "ik": {"chuai"}, "ir": {"chuan"}, "il": {"chuang"}, "iv": {"chui"}, "iy": {"chun"}, "io": {"chuo"},
	// j - ji series
	"ji": {"ji"}, "jx": {"jia"}, "jm": {"jian"},
	"jl": {"jiang"}, "jn": {"jiao"},
	"jp": {"jie"},
	"jb": {"jin"}, "jk": {"jing"}, "jq": {"jiu"},
	"js": {"jiong"},
	"jv": {"ju"}, "jr": {"juan"}, "jt": {"jue"}, "jy": {"jun"},
	// k - ka series
	"ka": {"ka"}, "kd": {"kai"}, "kj": {"kan"}, "kh": {"kang"}, "kc": {"kao"},
	"ke": {"ke"}, "kw": {"kei"}, "kf": {"ken"}, "kg": {"keng"},
	"ks": {"kong"}, "kz": {"kou"},
	"ku": {"ku"}, "kx": {"kua"}, "kk": {"kuai"}, "kr": {"kuan"}, "kl": {"kuang"}, "kv": {"kui"}, "ky": {"kun"}, "ko": {"kuo"},
	// l - la series
	"la": {"la"}, "ld": {"lai"}, "lj": {"lan"}, "lh": {"lang"}, "lc": {"lao"},
	"le": {"le"}, "lw": {"lei"}, "lg": {"leng"},
	"li": {"li"}, "lx": {"lia"}, "lm": {"lian"},
	"ll": {"liang"}, "ln": {"liao"},
	"lp": {"lie"},
	"lb": {"lin"}, "lk": {"ling"}, "lq": {"liu"},
	"lo": {"luo"}, "ls": {"long"}, "lz": {"lou"},
	"lu": {"lu"}, "lv": {"lv"}, "lr": {"luan"}, "lt": {"lve"}, "ly": {"lun"},
	// m - ma series
	"ma": {"ma"}, "md": {"mai"}, "mj": {"man"}, "mh": {"mang"}, "mc": {"mao"},
	"me": {"me"}, "mw": {"mei"}, "mf": {"men"}, "mg": {"meng"},
	"mi": {"mi"}, "mm": {"mian"}, "mn": {"miao"}, "mp": {"mie"}, "mb": {"min"}, "mk": {"ming"},
	"mq": {"miu"}, "mo": {"mo"}, "mz": {"mou"}, "mu": {"mu"},
	// n - na series
	"na": {"na"}, "nd": {"nai"}, "nj": {"nan"}, "nh": {"nang"}, "nc": {"nao"},
	"ne": {"ne"}, "nw": {"nei"}, "nf": {"nen"}, "ng": {"neng"},
	"ni": {"ni"}, "nm": {"nian"},
	"nl": {"niang"}, "nn": {"niao"}, "np": {"nie"},
	"nb": {"nin"}, "nk": {"ning"}, "nq": {"niu"},
	"ns": {"nong"}, "nz": {"nou"},
	"nu": {"nu"}, "nv": {"nv"}, "nr": {"nuan"}, "nt": {"nve"}, "no": {"nuo"},
	// o - zero-initial
	"oo": {"o"}, "oz": {"ou"},
	// p - pa series
	"pa": {"pa"}, "pd": {"pai"}, "pj": {"pan"}, "ph": {"pang"}, "pc": {"pao"},
	"pw": {"pei"}, "pf": {"pen"}, "pg": {"peng"},
	"pi": {"pi"}, "pm": {"pian"}, "pn": {"piao"}, "pp": {"pie"}, "pb": {"pin"}, "pk": {"ping"},
	"po": {"po"}, "pz": {"pou"}, "pu": {"pu"},
	// q - qi series
	"qi": {"qi"}, "qx": {"qia"}, "qm": {"qian"},
	"ql": {"qiang"}, "qn": {"qiao"},
	"qp": {"qie"},
	"qb": {"qin"}, "qk": {"qing"}, "qq": {"qiu"},
	"qs": {"qiong"},
	"qv": {"qu"}, "qr": {"quan"}, "qt": {"que"}, "qy": {"qun"},
	// r - ra series
	"rj": {"ran"}, "rh": {"rang"}, "rc": {"rao"},
	"re": {"re"}, "rf": {"ren"}, "rg": {"reng"}, "ri": {"ri"},
	"rs": {"rong"}, "rz": {"rou"},
	"ru": {"ru"}, "rr": {"ruan"}, "rv": {"rui"}, "ry": {"run"}, "ro": {"ruo"},
	// s - sa series
	"sa": {"sa"}, "sd": {"sai"}, "sj": {"san"}, "sh": {"sang"}, "sc": {"sao"},
	"se": {"se"}, "sf": {"sen"}, "sg": {"seng"}, "si": {"si"},
	"ss": {"song"}, "sz": {"sou"},
	"su": {"su"}, "sr": {"suan"}, "sv": {"sui"}, "sy": {"sun"}, "so": {"suo"},
	// t - ta series
	"ta": {"ta"}, "td": {"tai"}, "tj": {"tan"}, "th": {"tang"}, "tc": {"tao"},
	"te": {"te"}, "tg": {"teng"},
	"ti": {"ti"}, "tm": {"tian"}, "tn": {"tiao"}, "tp": {"tie"},
	"tk": {"ting"},
	"ts": {"tong"}, "tz": {"tou"},
	"tu": {"tu"}, "tr": {"tuan"}, "tv": {"tui"}, "ty": {"tun"}, "to": {"tuo"},
	// u (sh)
	"ua": {"sha"}, "ud": {"shai"}, "uj": {"shan"}, "uh": {"shang"}, "uc": {"shao"},
	"ue": {"she"}, "uw": {"shei"}, "uf": {"shen"}, "ug": {"sheng"},
	"ui": {"shi"}, "uz": {"shou"},
	"uu": {"shu"}, "ux": {"shua"}, "uk": {"shuai"}, "ur": {"shuan"}, "ul": {"shuang"}, "uv": {"shui"}, "uy": {"shun"}, "uo": {"shuo"},
	// v (zh)
	"va": {"zha"}, "vd": {"zhai"}, "vj": {"zhan"}, "vh": {"zhang"}, "vc": {"zhao"},
	"ve": {"zhe"}, "vf": {"zhen"}, "vg": {"zheng"},
	"vi": {"zhi"}, "vs": {"zhong"}, "vz": {"zhou"},
	"vu": {"zhu"}, "vx": {"zhua"}, "vk": {"zhuai"}, "vr": {"zhuan"}, "vl": {"zhuang"}, "vv": {"zhui"}, "vy": {"zhun"}, "vo": {"zhuo"},
	// w - wa series
	"wa": {"wa"}, "wd": {"wai"}, "wj": {"wan"}, "wh": {"wang"},
	"ww": {"wei"}, "wf": {"wen"}, "wg": {"weng"},
	"wo": {"wo"}, "wu": {"wu"},
	// x - xi series
	"xi": {"xi"}, "xx": {"xia"}, "xm": {"xian"},
	"xl": {"xiang"}, "xn": {"xiao"},
	"xp": {"xie"},
	"xb": {"xin"}, "xk": {"xing"}, "xq": {"xiu"},
	"xs": {"xiong"},
	"xv": {"xu"}, "xr": {"xuan"}, "xt": {"xue"}, "xy": {"xun"},
	// y - ya series
	"ya": {"ya"}, "yj": {"yan"}, "yh": {"yang"}, "yc": {"yao"},
	"ye": {"ye"}, "yi": {"yi"},
	"yb": {"yin"},
	"yk": {"ying"}, "yy": {"yun"}, "yo": {"yo"},
	"ys": {"yong"}, "yz": {"you"},
	"yv": {"yu"}, "yr": {"yuan"}, "yt": {"yue"},
	// z - za series
	"za": {"za"}, "zd": {"zai"}, "zj": {"zan"}, "zh": {"zang"}, "zc": {"zao"},
	"ze": {"ze"}, "zw": {"zei"}, "zf": {"zen"}, "zg": {"zeng"}, "zi": {"zi"},
	"zs": {"zong"}, "zz": {"zou"},
	"zu": {"zu"}, "zr": {"zuan"}, "zv": {"zui"}, "zy": {"zun"}, "zo": {"zuo"},
}

// XiaoheSpeller converts Xiaohe (小鹤双拼) codes to pinyin syllables.
type XiaoheSpeller struct{}

// NewXiaoheSpeller creates a new XiaoheSpeller.
func NewXiaoheSpeller() *XiaoheSpeller {
	return &XiaoheSpeller{}
}

// Name returns the speller name.
func (s *XiaoheSpeller) Name() string {
	return "xiaohe"
}

// ToPinyin converts a Xiaohe code string to a list of pinyin syllables.
// The input should be an even-length string of 2-character code pairs.
func (s *XiaoheSpeller) ToPinyin(code string) []string {
	if code == "" {
		return nil
	}
	var result []string
	for i := 0; i < len(code); i += 2 {
		if i+1 >= len(code) {
			break
		}
		chunk := code[i : i+2]
		if syllables, ok := xiaoheLookup[chunk]; ok {
			result = append(result, syllables...)
		} else {
			result = append(result, chunk)
		}
	}
	return result
}

// FullPinyinSpeller passes through full pinyin input as-is.
type FullPinyinSpeller struct{}

// NewFullPinyinSpeller creates a new FullPinyinSpeller.
func NewFullPinyinSpeller() *FullPinyinSpeller {
	return &FullPinyinSpeller{}
}

// Name returns the speller name.
func (s *FullPinyinSpeller) Name() string {
	return "fullpin"
}

// ToPinyin returns the input as a single-element slice.
func (s *FullPinyinSpeller) ToPinyin(code string) []string {
	if code == "" {
		return nil
	}
	return []string{code}
}
