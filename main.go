package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const (
	TYPE_START          = "start_kyoku"
	TYPE_TSUMO          = "tsumo"
	TYPE_DAHAI          = "dahai"
	TYPE_REACH          = "reach"
	TYPE_REACH_ACCEPTED = "reach_accepted"
	TYPE_CHI            = "chi"
	TYPE_PON            = "pon"
)

const BAD_PLAY_THRETHOLD = 8500

type Result [][]StartOrAction

type StartOrAction struct {
	EndNum    int   `json:"end_num"`
	Info      Info  `json:"info"`
	DahaiPred []int `json:"dahai_pred"`
}

type Info struct {
	Msg Msg `json:"msg"`
}

type Msg struct {
	Type      string     `json:"type"`
	Actor     int        `json:"actor"`
	Tehais    [][]string `json:"tehais"`
	Kyoku     int        `json:"kyoku"`
	Bakaze    string     `json:"bakaze"`
	Pai       string     `json:"pai"`
	RealDahai string     `json:"real_dahai"`
	PredDahai string     `json:"pred_dahai"`
}

type TehaiMap map[int][]string

func (t TehaiMap) changeTehai(actor int, tsumoPai, dahai string) {
	tehai := t[actor]
	tehai = append(tehai, tsumoPai)
	tehai = remove(tehai, dahai)
	t[actor] = tehai
}

type nagaJudge struct {
	point        float32
	judgeCount   int
	nagaRate     float32
	badPlayCount int
}

type actorNagaMap map[int]*nagaJudge

func (a actorNagaMap) culcNagaRate() {
	for actor := 0; actor <= 3; actor++ {
		a[actor].nagaRate = a[actor].point / float32(a[actor].judgeCount)
		log.Printf("actor: %v, nagaJudgeCount: %v, point: %v,nagaRate: %v, badPlayCount: %v", actor, a[actor].judgeCount, a[actor].point, a[actor].nagaRate, a[actor].badPlayCount)
	}
}

func main() {
	actorNagaMap := actorNagaMap{
		0: {0, 0, 0, 0},
		1: {0, 0, 0, 0},
		2: {0, 0, 0, 0},
		3: {0, 0, 0, 0},
	}
	realPaiNagaPaiIndexMap := getRealPaiNagaPaiIndexMap()
	var result Result
	playerTehaiMap := make(TehaiMap)
	resultRaw, err := ioutil.ReadFile("./target.json")
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	json.Unmarshal(resultRaw, &result)
	var kyokuIndex int
	for kyokuCount := len(result); kyokuCount > 0; kyokuCount-- {
		startOrActionLeftCount := len(result[kyokuIndex])
		startKyoku := result[kyokuIndex][0]
		for actor := 0; actor <= 3; actor++ {
			playerTehaiMap[actor] = startKyoku.Info.Msg.Tehais[actor]
		}
		startOrActionLeftCount--
		for actionIndex := 0; startOrActionLeftCount > 0; startOrActionLeftCount-- {
			actionIndex++
			action := result[kyokuIndex][actionIndex]
			actor := action.Info.Msg.Actor
			if action.Info.Msg.Type == TYPE_TSUMO {
				// 手配交換
				// playerTehaiMap.changeTehai(actor, action.Info.Msg.Pai, action.Info.Msg.RealDahai)
				actorNagaMap[actor].judgeCount++
				// NAG推奨打廃できてない場合は、推奨レート次第で加算ポイントを変える
				if action.Info.Msg.RealDahai != action.Info.Msg.PredDahai {
					realDahaiNagaPredRate := action.DahaiPred[realPaiNagaPaiIndexMap[action.Info.Msg.RealDahai]]
					predDahaiNagaPredRate := action.DahaiPred[realPaiNagaPaiIndexMap[action.Info.Msg.PredDahai]]
					if !(realDahaiNagaPredRate > 0) || !(predDahaiNagaPredRate > 0) {
						continue
					}
					if predDahaiNagaPredRate-realDahaiNagaPredRate > BAD_PLAY_THRETHOLD {
						if actor == 2 {
							// 悪手を出力してみる
							log.Printf("!!badPlay!! actor: %v, playerChoice: %v, playerChoicePredRate: %v nagaChoice: %v nagaChoicePredRate: %v", actor, action.Info.Msg.RealDahai, realDahaiNagaPredRate, action.Info.Msg.PredDahai, predDahaiNagaPredRate)
						}
						actorNagaMap[actor].badPlayCount++
					}
					actorNagaMap[actor].point = actorNagaMap[actor].point + float32(realDahaiNagaPredRate/predDahaiNagaPredRate)
					continue
				}
				actorNagaMap[actor].point++
			}

		}
		kyokuIndex++
	}
	actorNagaMap.culcNagaRate()
}

func getRealPaiNagaPaiIndexMap() map[string]int {
	realPaiNagaPaiIndexMap := make(map[string]int)
	realPaiNagaPaiIndexMap["1m"] = 0
	realPaiNagaPaiIndexMap["2m"] = 1
	realPaiNagaPaiIndexMap["3m"] = 2
	realPaiNagaPaiIndexMap["4m"] = 3
	realPaiNagaPaiIndexMap["5m"] = 4
	realPaiNagaPaiIndexMap["6m"] = 5
	realPaiNagaPaiIndexMap["7m"] = 6
	realPaiNagaPaiIndexMap["8m"] = 7
	realPaiNagaPaiIndexMap["9m"] = 8
	realPaiNagaPaiIndexMap["1p"] = 9
	realPaiNagaPaiIndexMap["2p"] = 10
	realPaiNagaPaiIndexMap["3p"] = 11
	realPaiNagaPaiIndexMap["4p"] = 12
	realPaiNagaPaiIndexMap["5p"] = 13
	realPaiNagaPaiIndexMap["6p"] = 14
	realPaiNagaPaiIndexMap["7p"] = 15
	realPaiNagaPaiIndexMap["8p"] = 16
	realPaiNagaPaiIndexMap["9p"] = 17
	realPaiNagaPaiIndexMap["1s"] = 18
	realPaiNagaPaiIndexMap["2s"] = 19
	realPaiNagaPaiIndexMap["3s"] = 20
	realPaiNagaPaiIndexMap["4s"] = 21
	realPaiNagaPaiIndexMap["5s"] = 22
	realPaiNagaPaiIndexMap["6s"] = 23
	realPaiNagaPaiIndexMap["7s"] = 24
	realPaiNagaPaiIndexMap["8s"] = 25
	realPaiNagaPaiIndexMap["9s"] = 26
	realPaiNagaPaiIndexMap["E"] = 27
	realPaiNagaPaiIndexMap["S"] = 28
	realPaiNagaPaiIndexMap["W"] = 29
	realPaiNagaPaiIndexMap["N"] = 30
	realPaiNagaPaiIndexMap["P"] = 31
	realPaiNagaPaiIndexMap["F"] = 32
	realPaiNagaPaiIndexMap["C"] = 33
	return realPaiNagaPaiIndexMap
}

func remove(strings []string, search string) []string {
	result := []string{}
	var isAlreadyRemoved bool
	for _, v := range strings {
		if v != search && !isAlreadyRemoved {
			isAlreadyRemoved = true
			continue
		}
		result = append(result, v)
	}
	return result
}
