package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olivere/elastic"
)

func main() {
	var (
		filePath = "./stat.log"
		reader   = StatLogReader{
			filePath: filePath,
			skipRows: 2,
		}
		ctx = context.Background()
		err error
	)
	client, err := elastic.NewClient(
		elastic.SetURL("http://192.168.33.102:9200"),
		elastic.SetHealthcheck(false),
		elastic.SetSniff(false),
	)
	if err != nil {
		panic(err)
	}
	callback := Callback{
		OnReadLine: func(values []string) {
			stat := toStat(values)
			resp, err := client.Index().Index(
				"logstash-201808",
			).Type(
				"dashboard",
			).Id(
				fmt.Sprintf("%s%d", stat.HostName, stat.Timestamp.Unix()),
			).BodyJson(
				stat,
			).Do(ctx)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%v", resp)
			fmt.Println("")
		},
	}
	err = reader.Read(callback)
	if err != nil {
		panic(err)
	}
}

type StatLogReader struct {
	filePath string
	skipRows int
}

type Callback struct {
	OnReadLine func(values []string)
}

func (r *StatLogReader) Read(cb Callback) error {
	var file *os.File
	file, err := os.Open(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	pos := 0
	reader := bufio.NewReaderSize(file, 4096)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		pos++
		if pos <= r.skipRows {
			continue
		}
		cb.OnReadLine(parseLine(line))
	}
	return nil
}

type Stat struct {
	HostName  string
	Timestamp time.Time
	Procs     Procs
	Memory    Memory
	Swap      Swap
	IO        IO
	System    System
	CPU       CPU
}

func (s *Stat) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Timestamp    time.Time `json:"@timestamp"`
		HostName     string    `json:"hostname"`
		Procs_R      int       `json:"proc_r"`
		Procs_B      int       `json:"proc_b"`
		Memory_Swapd int       `json:"mem_swpd"`
		Memory_Free  int       `json:"mem_free"`
		Memory_Buff  int       `json:"mem_buff"`
		Memory_Cache int       `json:"mem_cache"`
		Swap_SI      int       `json:"swp_si"`
		Swap_SO      int       `json:"swp_so"`
		IO_BI        int       `json:"io_bi"`
		IO_BO        int       `json:"io_bo"`
		System_IN    int       `json:"system_in"`
		System_CS    int       `json:"system_cs"`
		CPU_US       int       `json:"cpu_us"`
		CPU_SY       int       `json:"cpu_sy"`
		CPU_ID       int       `json:"cpu_id"`
		CPU_WA       int       `json:"cpu_wa"`
		CPU_ST       int       `json:"cpu_st"`
	}{
		Timestamp:    s.Timestamp,
		HostName:     s.HostName,
		Procs_R:      s.Procs.R,
		Procs_B:      s.Procs.B,
		Memory_Swapd: s.Memory.Swpd,
		Memory_Free:  s.Memory.Free,
		Memory_Buff:  s.Memory.Buff,
		Memory_Cache: s.Memory.Cache,
		Swap_SI:      s.Swap.SI,
		Swap_SO:      s.Swap.SO,
		IO_BI:        s.IO.BI,
		IO_BO:        s.IO.BO,
		System_IN:    s.System.IN,
		System_CS:    s.System.CS,
		CPU_US:       s.CPU.US,
		CPU_SY:       s.CPU.SY,
		CPU_ID:       s.CPU.ID,
		CPU_WA:       s.CPU.WA,
		CPU_ST:       s.CPU.ST,
	})
}

func toStat(values []string) *Stat {
	return &Stat{
		HostName:  values[0],
		Timestamp: mustParseTime(time.RFC3339, values[1]),
		Procs: Procs{
			R: mustParseInt(values[2]),
			B: mustParseInt(values[3]),
		},
		Memory: Memory{
			Swpd:  mustParseInt(values[4]),
			Free:  mustParseInt(values[5]),
			Buff:  mustParseInt(values[6]),
			Cache: mustParseInt(values[7]),
		},
		Swap: Swap{
			SI: mustParseInt(values[8]),
			SO: mustParseInt(values[9]),
		},
		IO: IO{
			BI: mustParseInt(values[10]),
			BO: mustParseInt(values[11]),
		},
		System: System{
			IN: mustParseInt(values[12]),
			CS: mustParseInt(values[13]),
		},
		CPU: CPU{
			US: mustParseInt(values[14]),
			SY: mustParseInt(values[15]),
			ID: mustParseInt(values[16]),
			WA: mustParseInt(values[17]),
			ST: mustParseInt(values[18]),
		},
	}
}

type (
	Procs struct {
		R int // 実行待ちプロセス数
		B int // 割り込み不可能なスリープ中のプロセス数
	}
	Memory struct {
		Swpd  int // バーチャルメモリの使用量
		Free  int // 空きメモリ量
		Buff  int // バッファに使われてるメモリ量
		Cache int // キャッシュに使われているメモリ量
	}
	Swap struct {
		SI int // 秒あたりのスワップイン量
		SO int // 秒あたりのスワップアウト量
	}
	IO struct {
		BI int // 秒あたりのブロックデバイスから受け取ったブロック数
		BO int // 秒あたりのブロックデバイスに送ったブロック数
	}
	System struct {
		IN int // 秒あたりの割り込み回数
		CS int // 秒あたりのコンテキストスイッチの回数
	}
	CPU struct {
		US int // カーネル以外のコードでかかっている時間
		SY int // カーネルコードでかかっている時間
		ID int // アイドルタイム
		WA int // IO待ち時間
		ST int // 要求したがCPUリソースを割り当ててもらえなかった時間
	}
)

func mustParseTime(layout, value string) time.Time {
	ret, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return ret
}

func mustParseInt(value string) int {
	ret, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		panic(err)
	}
	return int(ret)
}

func parseLine(line string) (values []string) {
	for _, v := range strings.Split(line, " ") {
		cnvValue := strings.Replace(v, " ", "", -1)
		cnvValue = strings.Replace(cnvValue, "\n", "", -1)
		if len(cnvValue) == 0 {
			continue
		}
		values = append(values, cnvValue)
	}
	return
}

