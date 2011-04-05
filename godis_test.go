package godis

import (
    "testing"
    "bytes"
    "bufio"
    "os"
    "time"
    "log"
)

func compareReply(t *testing.T, name string, a, b *Reply) {
    if a.Err != nil && b.Err == nil {
        t.Fatalf("'%s': expected error `%v`", name, a.Err)
    } else if b.Err != a.Err {
        t.Fatalf("'%s': expected %s got %v", name, a.Err, b.Err)
    } else if b.Elem != nil {
        for i, c := range a.Elem {
            if c != b.Elem[i] {
                t.Errorf("'%s': expected %v got %v", name, b, a)
            }
        }
    } else if b.Elems != nil {
        for i, rep := range a.Elems {
            for j, e := range rep.Elem {
                if e != b.Elems[i].Elem[j] {
                    t.Errorf("expected %v got %v", b, a)
                    break
                }
            }
        }
    }
}

type simpleParserTest struct {
    in   string
    out  Reply
    name string
}

type redisReadWriter struct {
    writer *bufio.Writer
    reader *bufio.Reader
}

func dummyReadWriter(data string) *redisReadWriter {
    r := bufio.NewReader(bytes.NewBufferString(data))
    w := bufio.NewWriter(bytes.NewBufferString(data))
    return &redisReadWriter{w, r}
}

var simpleParserTests = []simpleParserTest{
    {"+OK\r\n", Reply{Elem: []byte("OK")}, "ok"},
    {"-ERR\r\n", Reply{Err: os.NewError("ERR")}, "err"},
    {":1\r\n", Reply{Elem: []byte("1")}, "num"},
    {"$3\r\nfoo\r\n", Reply{Elem: []byte("foo")}, "bulk"},
    {"$-1\r\n", Reply{}, "bulk-nil"},
    {"*-1\r\n", Reply{}, "multi-bulk-nil"},
}

func TestParser(t *testing.T) {
    for _, test := range simpleParserTests {
        rw := dummyReadWriter(test.in)
        r := parseResponse(rw.reader)
        compareReply(t, test.name, r, &test.out)
        t.Log(test.in, r, test.out)
    }
}

func s2MultiReply(ss ...string) []*Reply {
    var r = make([]*Reply, len(ss))
    for i := 0; i < len(ss); i++ {
        r[i] = &Reply{Elem: []byte(ss[i])}
    }
    return r
}

type SimpleSendTest struct {
    cmd  string
    args []string
    out  Reply
}

var simpleSendTests = []SimpleSendTest{
    {"FLUSHDB", []string{}, Reply{Elem: []byte("OK")}},
    {"SET", []string{"key", "foo"}, Reply{Elem: []byte("OK")}},
    {"EXISTS", []string{"key"}, Reply{Elem: []byte("1")}},
    {"GET", []string{"key"}, Reply{Elem: []byte("foo")}},
    {"RPUSH", []string{"list", "foo"}, Reply{Elem: []byte("1")}},
    {"RPUSH", []string{"list", "bar"}, Reply{Elem: []byte("2")}},
    {"LRANGE", []string{"list", "0", "2"}, Reply{Elems: s2MultiReply("foo", "bar")}},
    {"KEYS", []string{"list"}, Reply{Elems: s2MultiReply("list")}},
    {"GET", []string{"/dev/null"}, Reply{}},
}

func TestSimpleSend(t *testing.T) {
    c := New("", 0, "")
    for _, test := range simpleSendTests {
        r := SendStr(c, test.cmd, test.args...)
        compareReply(t, test.cmd, &test.out, r)
        t.Log(test.cmd, test.args)
        t.Logf("%q == %q\n", test.out, r)
    }
}

//func TestSimplePipe(t *testing.T) {
//    c := NewPipe("", 0, "")
//    
//    for _, test := range simpleSendTests {
//        r := SendStr(c, test.cmd, test.args...)
//        if r.Err != nil {
//            t.Error(test.cmd, r.Err, test.args)
//        }
//    }
//
//    for _, test := range simpleSendTests {
//        r := c.GetReply()
//        compareReply(t, test.cmd, &test.out, r)
//        t.Log(test.cmd, test.args)
//        t.Logf("%q == %q\n", test.out, r)
//    }
//}

func BenchmarkParsing(b *testing.B) {
    c := New("", 0, "")

    for i := 0; i < 1000; i++ {
        SendStr(c, "RPUSH", "list", "foo")
    }

    start := time.Nanoseconds()

    for i := 0; i < b.N; i++ {
        SendStr(c, "LRANGE", "list", "0", "50")
    }

    stop := time.Nanoseconds() - start

    log.Printf("time: %.2f\n", float32(stop/1.0e+6)/1000.0)
    Send(c, []byte("FLUSHDB"))
}

//func TestBenchmark(t *testing.T) {
//    c := New("", 0, "")
//    c.Send("FLUSHDB")
//    start := time.Nanoseconds()
//    n := 2000000
//
//    a, b := []byte("zrs"), []byte("hi")
//    for i := 0; i < n; i++ {
//        c.Send("RPUSH", a, b)
//    }
//
//    //c.Del("zrs")
//    stop := time.Nanoseconds() - start
//
//    ti := float32(stop / 1.0e+6) / 1000.0
//    fmt.Fprintf(os.Stdout, "godis: %.2f %.2f per/s\n", ti, float32(n) / ti)
//}
