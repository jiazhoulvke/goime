package engine

// Speller defines the interface for converting input codes to pinyin syllables.
type Speller interface {
	Name() string
	ToPinyin(code string) []string
}

// xiaoheLookup maps 2-character Xiaohe codes to their pinyin syllable(s).
var xiaoheLookup = map[string][]string{
	// a
	"aa": {"a"},
	// zero-initial a series
	"ad": {"ai"}, "aj": {"an"}, "ah": {"ang"}, "ak": {"ao"},
	// b - ba series
	"ba": {"ba"}, "bd": {"bai"}, "bj": {"ban"}, "bh": {"bang"}, "bk": {"bao"},
	"bw": {"bei"}, "bf": {"ben"}, "bg": {"beng"},
	"bi": {"bi"}, "bm": {"bian"}, "bc": {"biao"}, "bp": {"bie"}, "bl": {"bin"}, "by": {"bing"},
	"bo": {"bo"}, "bu": {"bu"},
	// c - ca series
	"ca": {"ca"}, "cd": {"cai"}, "cj": {"can"}, "ch": {"cang"}, "ck": {"cao"},
	"ce": {"ce"}, "cf": {"cen"}, "cg": {"ceng"}, "ci": {"ci"},
	"cs": {"cong"}, "cz": {"cou"}, "cu": {"cu"},
	"cr": {"cuan"}, "cv": {"cui"}, "cy": {"cun"}, "co": {"cuo"},
	// d - da series
	"da": {"da"}, "dd": {"dai"}, "dj": {"dan"}, "dh": {"dang"}, "dk": {"dao"},
	"de": {"de"}, "dw": {"dei"}, "dg": {"deng"},
	"di": {"di"}, "db": {"dia"}, "dm": {"dian"}, "dc": {"diao"}, "dp": {"die"},
	"dy": {"ding", "dun"}, "dq": {"diu"},
	"ds": {"dong"}, "dz": {"dou"},
	"du": {"du"}, "dr": {"duan"}, "dv": {"dui"}, "do": {"duo"},
	// e - zero-initial
	"ee": {"e"}, "ew": {"ei"}, "eb": {"en"}, "eg": {"eng"}, "er": {"er"},
	// f - fa series
	"fa": {"fa"}, "fj": {"fan"}, "fh": {"fang"}, "fk": {"fao"},
	"fw": {"fei"}, "ff": {"fen"}, "fg": {"feng"},
	"fi": {"fi"}, "fo": {"fo"}, "fz": {"fou"}, "fu": {"fu"},
	// g - ga series
	"ga": {"ga"}, "gd": {"gai"}, "gj": {"gan"}, "gh": {"gang"}, "gk": {"gao"},
	"ge": {"ge"}, "gi": {"gei"}, "gf": {"gen"}, "gg": {"geng"},
	"gs": {"gong"}, "gz": {"gou"},
	"gu": {"gu"}, "gb": {"gua"}, "gx": {"guai"}, "gr": {"guan"}, "gl": {"guang"}, "gv": {"gui"}, "gy": {"gun"}, "go": {"guo"},
	// h - ha series
	"ha": {"ha"}, "hd": {"hai"}, "hj": {"han"}, "hh": {"hang"}, "hk": {"hao"},
	"he": {"he"}, "hi": {"hei"}, "hf": {"hen"}, "hg": {"heng"},
	"hs": {"hong"}, "hz": {"hou"},
	"hu": {"hu"}, "hb": {"hua"}, "hx": {"huai"}, "hr": {"huan"}, "hl": {"huang"}, "hv": {"hui"}, "hy": {"hun"}, "ho": {"huo"},
	// i (ch)
	"ia": {"cha"}, "id": {"chai"}, "ij": {"chan"}, "ih": {"chang"}, "ik": {"chao"},
	"ie": {"che"}, "ib": {"chen"}, "ig": {"cheng"},
	"ii": {"chi"}, "is": {"chong"}, "iz": {"chou"},
	"iu": {"chu"}, "ic": {"chua"}, "ix": {"chuai"}, "ir": {"chuan"}, "il": {"chuang"}, "iv": {"chui"}, "iy": {"chun"}, "io": {"chuo"},
	// j - ji series
	"ji": {"ji"}, "jb": {"jia"}, "jm": {"jian"},
	"jl": {"jiang", "jin"}, "jc": {"jiao"},
	"jp": {"jie"},
	"jy": {"jing", "jun"}, "js": {"jiong"}, "jq": {"jiu"},
	"jv": {"ju"}, "jr": {"juan"}, "jt": {"jue"},
	// k - ka series
	"ka": {"ka"}, "kd": {"kai"}, "kj": {"kan"}, "kh": {"kang"}, "kk": {"kao"},
	"ke": {"ke"}, "ki": {"kei"}, "kf": {"ken"}, "kg": {"keng"},
	"ks": {"kong"}, "kz": {"kou"},
	"ku": {"ku"}, "kb": {"kua"}, "kx": {"kuai"}, "kr": {"kuan"}, "kl": {"kuang"}, "kv": {"kui"}, "ky": {"kun"}, "ko": {"kuo"},
	// l - la series
	"la": {"la"}, "ld": {"lai"}, "lj": {"lan"}, "lh": {"lang"}, "lk": {"lao"},
	"le": {"le"}, "lw": {"lei"}, "lg": {"leng"},
	"li": {"li"}, "lb": {"lia"}, "lm": {"lian"},
	"ll": {"liang", "lin"}, "lc": {"liao"},
	"lp": {"lie"},
	"ly": {"ling", "lun"}, "lq": {"liu"},
	"lo": {"lo", "luo"}, "ls": {"long"}, "lz": {"lou"},
	"lu": {"lu"}, "lv": {"lv"}, "lr": {"luan"}, "lt": {"lve"},
	// m - ma series
	"ma": {"ma"}, "md": {"mai"}, "mj": {"man"}, "mh": {"mang"}, "mk": {"mao"},
	"me": {"me"}, "mw": {"mei"}, "mf": {"men"}, "mg": {"meng"},
	"mi": {"mi"}, "mm": {"mian"}, "mc": {"miao"}, "mp": {"mie"}, "ml": {"min"}, "my": {"ming"},
	"miu": {"miu"}, "mo": {"mo"}, "mz": {"mou"}, "mu": {"mu"},
	// n - na series
	"na": {"na"}, "nd": {"nai"}, "nj": {"nan"}, "nh": {"nang"}, "nk": {"nao"},
	"ne": {"ne"}, "nw": {"nei"}, "nf": {"nen"}, "ng": {"neng"},
	"ni": {"ni"}, "nm": {"nian"},
	"nl": {"niang", "nin"}, "nc": {"niao"}, "np": {"nie"},
	"nn": {"nin"}, "ny": {"ning"}, "nq": {"niu"},
	"ns": {"nong"}, "nz": {"nou"},
	"nu": {"nu"}, "nv": {"nv"}, "nr": {"nuan"}, "nt": {"nve"}, "no": {"nuo"},
	// p - pa series
	"pa": {"pa"}, "pd": {"pai"}, "pj": {"pan"}, "ph": {"pang"}, "pk": {"pao"},
	"pw": {"pei"}, "pf": {"pen"}, "pg": {"peng"},
	"pi": {"pi"}, "pm": {"pian"}, "pc": {"piao"}, "pp": {"pie"}, "pl": {"pin"}, "py": {"ping"},
	"po": {"po"}, "pz": {"pou"}, "pu": {"pu"},
	// q - qi series
	"qi": {"qi"}, "qb": {"qia"}, "qm": {"qian"},
	"ql": {"qiang", "qin"}, "qc": {"qiao"},
	"qp": {"qie"},
	"qy": {"qing", "qun"}, "qs": {"qiong"}, "qq": {"qiu"},
	"qv": {"qu"}, "qr": {"quan"}, "qt": {"que"},
	// r - ra series
	"rj": {"ran"}, "rh": {"rang"}, "rk": {"rao"},
	"re": {"re"}, "rf": {"ren"}, "rg": {"reng"}, "ri": {"ri"},
	"rs": {"rong"}, "rz": {"rou"},
	"ru": {"ru"}, "rr": {"ruan"}, "rv": {"rui"}, "ry": {"run"}, "ro": {"ruo"},
	// s - sa series
	"sa": {"sa"}, "sd": {"sai"}, "sj": {"san"}, "sh": {"sang"}, "sk": {"sao"},
	"se": {"se"}, "sf": {"sen"}, "sg": {"seng"}, "si": {"si"},
	"ss": {"song"}, "sz": {"sou"},
	"su": {"su"}, "sr": {"suan"}, "sv": {"sui"}, "sy": {"sun"}, "so": {"suo"},
	// t - ta series
	"ta": {"ta"}, "td": {"tai"}, "tj": {"tan"}, "th": {"tang"}, "tk": {"tao"},
	"te": {"te"}, "tg": {"teng"},
	"ti": {"ti"}, "tm": {"tian"}, "tc": {"tiao"}, "tp": {"tie"},
	"ty": {"ting", "tun"},
	"ts": {"tong"}, "tz": {"tou"},
	"tu": {"tu"}, "tr": {"tuan"}, "tv": {"tui"}, "to": {"tuo"},
	// u (sh)
	"ua": {"sha"}, "ud": {"shai"}, "uj": {"shan"}, "uh": {"shang"}, "uk": {"shao"},
	"ue": {"she"}, "uw": {"shei"}, "ub": {"shen"}, "ug": {"sheng"},
	"ui": {"shi"}, "uz": {"shou"},
	"uu": {"shu"}, "uc": {"shua"}, "ux": {"shuai"}, "ur": {"shuan"}, "ul": {"shuang"}, "uv": {"shui"}, "uy": {"shun"}, "uo": {"shuo"},
	// v (zh)
	"va": {"zha"}, "vd": {"zhai"}, "vj": {"zhan"}, "vh": {"zhang"}, "vk": {"zhao"},
	"ve": {"zhe"}, "vf": {"zhen"}, "vg": {"zheng"},
	"vi": {"zhi"}, "vs": {"zhong"}, "vz": {"zhou"},
	"vu": {"zhu"}, "vc": {"zhua"}, "vx": {"zhuai"}, "vr": {"zhuan"}, "vl": {"zhuang"}, "vv": {"zhui"}, "vy": {"zhun"}, "vo": {"zhuo"},
	// w - wa series
	"wa": {"wa"}, "wd": {"wai"}, "wj": {"wan"}, "wh": {"wang"},
	"ww": {"wei"}, "wf": {"wen"}, "wg": {"weng"},
	"wo": {"wo"}, "wu": {"wu"},
	// x - xi series
	"xi": {"xi"}, "xb": {"xia"}, "xm": {"xian"},
	"xl": {"xiang", "xin"}, "xc": {"xiao"},
	"xp": {"xie"},
	"xy": {"xing", "xun"}, "xs": {"xiong"}, "xq": {"xiu"},
	"xv": {"xu"}, "xr": {"xuan"}, "xt": {"xue"},
	// y - ya series
	"ya": {"ya"}, "yj": {"yan"}, "yh": {"yang"}, "yk": {"yao"},
	"yc": {"yao"},
	// o - zero-initial
	"oo": {"o"}, "oz": {"ou"},
	"ye": {"ye"}, "yi": {"yi"},
	"yl": {"yin"},
	"yy": {"ying", "yun"}, "yo": {"yo"},
	"ys": {"yong"}, "yz": {"you"},
	"yv": {"yu"}, "yr": {"yuan"}, "yt": {"yue"},
	// z - za series
	"za": {"za"}, "zd": {"zai"}, "zj": {"zan"}, "zh": {"zang"}, "zk": {"zao"},
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
