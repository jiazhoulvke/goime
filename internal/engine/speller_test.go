package engine

import (
    "reflect"
    "testing"
)

func TestSpellerXiaohe(t *testing.T) {
    s := NewXiaoheSpeller()
    tests := []struct {
        input string
        want  []string
    }{
        {"uuru", []string{"shu", "ru"}},
        {"ni", []string{"ni"}},
        {"hf", []string{"hen"}},
        {"ji", []string{"ji"}},
        {"", nil},
        // zero-initial
        {"aa", []string{"a"}},
        {"ee", []string{"e"}},
        {"oo", []string{"o"}},
        // K = ing (b/p/m/f/d/t/n/l/j/q/x/y)
        {"bk", []string{"bing"}},
        {"pk", []string{"ping"}},
        {"mk", []string{"ming"}},
        {"dk", []string{"ding"}},
        {"tk", []string{"ting"}},
        {"nk", []string{"ning"}},
        {"lk", []string{"ling"}},
        {"jk", []string{"jing"}},
        {"qk", []string{"qing"}},
        {"xk", []string{"xing"}},
        {"yk", []string{"ying"}},
        // K = uai (g/k/h/zh/ch/sh/r)
        {"gk", []string{"guai"}},
        {"hk", []string{"huai"}},
        {"kk", []string{"kuai"}},
        {"vk", []string{"zhuai"}},
        {"ik", []string{"chuai"}},
        {"uk", []string{"shuai"}},
        // C = ao
        {"bc", []string{"bao"}},
        {"pc", []string{"pao"}},
        {"mc", []string{"mao"}},
        {"dc", []string{"dao"}},
        {"tc", []string{"tao"}},
        {"nc", []string{"nao"}},
        {"lc", []string{"lao"}},
        {"gc", []string{"gao"}},
        {"kc", []string{"kao"}},
        {"hc", []string{"hao"}},
        {"vc", []string{"zhao"}},
        {"ic", []string{"chao"}},
        {"uc", []string{"shao"}},
        {"rc", []string{"rao"}},
        {"zc", []string{"zao"}},
        {"cc", []string{"cao"}},
        {"sc", []string{"sao"}},
        {"yc", []string{"yao"}},
        // N = iao
        {"bn", []string{"biao"}},
        {"pn", []string{"piao"}},
        {"mn", []string{"miao"}},
        {"dn", []string{"diao"}},
        {"tn", []string{"tiao"}},
        {"nn", []string{"niao"}},
        {"ln", []string{"liao"}},
        {"jn", []string{"jiao"}},
        {"qn", []string{"qiao"}},
        {"xn", []string{"xiao"}},
        // B = in
        {"bb", []string{"bin"}},
        {"pb", []string{"pin"}},
        {"mb", []string{"min"}},
        {"nb", []string{"nin"}},
        {"lb", []string{"lin"}},
        {"jb", []string{"jin"}},
        {"qb", []string{"qin"}},
        {"xb", []string{"xin"}},
        {"yb", []string{"yin"}},
        // X = ia
        {"dx", []string{"dia"}},
        {"lx", []string{"lia"}},
        {"jx", []string{"jia"}},
        {"qx", []string{"qia"}},
        {"xx", []string{"xia"}},
        // X = ua
        {"gx", []string{"gua"}},
        {"kx", []string{"kua"}},
        {"hx", []string{"hua"}},
        {"vx", []string{"zhua"}},
        {"ix", []string{"chua"}},
        {"ux", []string{"shua"}},
        // Y = un
        {"dy", []string{"dun"}},
        {"ty", []string{"tun"}},
        {"ly", []string{"lun"}},
        {"gy", []string{"gun"}},
        {"ky", []string{"kun"}},
        {"hy", []string{"hun"}},
        {"vy", []string{"zhun"}},
        {"iy", []string{"chun"}},
        {"uy", []string{"shun"}},
        {"zy", []string{"zun"}},
        {"cy", []string{"cun"}},
        {"sy", []string{"sun"}},
        {"ry", []string{"run"}},
        {"yy", []string{"yun"}},
        {"jy", []string{"jun"}},
        {"qy", []string{"qun"}},
        {"xy", []string{"xun"}},
        // L = iang (j/q/x/n/l)
        {"jl", []string{"jiang"}},
        {"ql", []string{"qiang"}},
        {"xl", []string{"xiang"}},
        {"nl", []string{"niang"}},
        {"ll", []string{"liang"}},
        // L = uang (g/k/h/zh/ch/sh/r)
        {"gl", []string{"guang"}},
        {"kl", []string{"kuang"}},
        {"hl", []string{"huang"}},
        {"vl", []string{"zhuang"}},
        {"il", []string{"chuang"}},
        {"ul", []string{"shuang"}},
        // zero-initial ang
        {"ah", []string{"ang"}},
        // miu fix
        {"mq", []string{"miu"}},
        // no multi-syllable ambiguity
        {"lo", []string{"luo"}},
    }
    for _, tc := range tests {
        got := s.ToPinyin(tc.input)
        if !reflect.DeepEqual(got, tc.want) {
            t.Errorf("ToPinyin(%q) = %v, want %v", tc.input, got, tc.want)
        }
    }
}

func TestSpellerFullPinyin(t *testing.T) {
    s := NewFullPinyinSpeller()
    tests := []struct {
        input string
        want  []string
    }{
        {"shuru", []string{"shuru"}},
        {"ni", []string{"ni"}},
        {"hao", []string{"hao"}},
        {"", nil},
    }
    for _, tc := range tests {
        got := s.ToPinyin(tc.input)
        if !reflect.DeepEqual(got, tc.want) {
            t.Errorf("ToPinyin(%q) = %v, want %v", tc.input, got, tc.want)
        }
    }
}

func TestSpellerInterface(t *testing.T) {
    var s Speller = NewXiaoheSpeller()
    if s.Name() != "xiaohe" {
        t.Errorf("expected xiaohe, got %s", s.Name())
    }
    s = NewFullPinyinSpeller()
    if s.Name() != "fullpin" {
        t.Errorf("expected fullpin, got %s", s.Name())
    }
}
