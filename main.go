package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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

const (
	RECOMMEND_THRETHOLD          = 5000
	STRONGLY_RECOMMEND_THRETHOLD = 9000
)

type Result [][]StartOrAction

type StartOrAction struct {
	EndNum    int                     `json:"end_num"`
	Info      Info                    `json:"info"`
	DahaiPred []float32               `json:"dahai_pred"`
	Reach     float32                 `json:"reach"`
	Huro      map[int]map[int]float32 `json:"huro"`
}

type Info struct {
	Msg Msg `json:"msg"`
}

type Msg struct {
	Type       string     `json:"type"`
	Actor      int        `json:"actor"`
	Tehais     [][]string `json:"tehais"`
	Kyoku      int        `json:"kyoku"`
	Bakaze     string     `json:"bakaze"`
	Pai        string     `json:"pai"`
	RealDahai  string     `json:"real_dahai"`
	PredDahai  string     `json:"pred_dahai"`
	LeftHaiNum int        `json:"left_hai_num"`
}

type TehaiMap map[int][]string

func (t TehaiMap) changeTehai(actor int, tsumoPai, dahai string) {
	tehai := t[actor]
	tehai = append(tehai, tsumoPai)
	tehai = remove(tehai, dahai)
	t[actor] = tehai
}

type nagaJudge struct {
	judgeCount     int
	matchCount     int
	unMatchCount   int
	BadPlayCount   int
	minusValueList []float32
}

type nextHuroRecommend struct {
	nextShouldHuro bool
	rate           float32
	pattern        string
}

type nextReachRecommend struct {
	nextShouldReach bool
	rate            float32
}

type actorNagaMap map[int]*nagaJudge

func (a actorNagaMap) culcNagaValue() {
	for actor := 0; actor <= 3; actor++ {
		var minusSum float32
		for _, v := range a[actor].minusValueList {
			minusSum += v
		}
		nagaValue := 100 - minusSum/float32(len(a[actor].minusValueList)*100)
		log.Printf("actor: %v,nagaJudgeCount: %v,matchCount: %v,unMatchCount: %v,matchRate: %.2f,nagaValue: %.3f, BadPlayCount: %v", actor, a[actor].judgeCount, a[actor].matchCount, a[actor].unMatchCount, 100*float32(a[actor].matchCount)/float32(a[actor].judgeCount), nagaValue, a[actor].BadPlayCount)
	}
}

func main() {
	actorNagaMap := actorNagaMap{
		0: {0, 0, 0, 0, []float32{}},
		1: {0, 0, 0, 0, []float32{}},
		2: {0, 0, 0, 0, []float32{}},
		3: {0, 0, 0, 0, []float32{}},
	}
	var selectActor int
	flag.IntVar(&selectActor, "actor", 0, "log出力したいactorを選択")
	flag.Parse()
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
		bakaze := startKyoku.Info.Msg.Bakaze
		kyoku := startKyoku.Info.Msg.Kyoku
		startOrActionLeftCount--
		nextReachRecommend := map[int]*nextReachRecommend{
			0: {false, 0},
			1: {false, 0},
			2: {false, 0},
			3: {false, 0},
		}
		nextHuroRecommend := map[int]*nextHuroRecommend{
			0: {false, 0, ""},
			1: {false, 0, ""},
			2: {false, 0, ""},
			3: {false, 0, ""},
		}
		for actionIndex := 0; startOrActionLeftCount > 0; startOrActionLeftCount-- {
			actionIndex++
			action := result[kyokuIndex][actionIndex]
			actor := action.Info.Msg.Actor

			//// リーチ判断判定ゾーン
			//if nextReachRecommend[actor].nextShouldReach && action.Info.Msg.Type == TYPE_REACH {
			//	actorNagaMap[actor].judgeCount++
			//	actorNagaMap[actor].matchCount++
			//	actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, 0)
			//	nextReachRecommend[actor].nextShouldReach = false
			//	nextReachRecommend[actor].rate = 0
			//}
			//if nextReachRecommend[actor].nextShouldReach && action.Info.Msg.Type != TYPE_REACH {
			//	actorNagaMap[actor].judgeCount++
			//	actorNagaMap[actor].unMatchCount++
			//	actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, nextReachRecommend[actor].rate)
			//	if nextReachRecommend[actor].rate > BAD_PLAY_THRETHOLD {
			//		// 悪手を出力してみる
			//		if actor == selectActor {
			//			log.Printf("!!BADPLAYYYYYY 鉄Reach!! actor: %v, bakaze: %v, kyoku: %v,leftHiNum: %v, reachPredRate: %v", actor, bakaze, kyoku, action.Info.Msg.LeftHaiNum, nextReachRecommend[actor].rate)
			//		}
			//		actorNagaMap[actor].BadPlayCount++
			//		nextReachRecommend[actor].nextShouldReach = false
			//		nextReachRecommend[actor].rate = 0
			//	} else {
			//		if actor == selectActor {
			//			log.Printf("Reach推奨 actor: %v, bakaze: %v, kyoku: %v,leftHiNum: %v, reachPredRate: %v", actor, bakaze, kyoku, action.Info.Msg.LeftHaiNum, nextReachRecommend[actor].rate)
			//		}
			//		nextReachRecommend[actor].nextShouldReach = false
			//		nextReachRecommend[actor].rate = 0
			//	}
			//}
			//if !nextReachRecommend[actor].nextShouldReach && action.Info.Msg.Type == TYPE_REACH {
			//	actorNagaMap[actor].judgeCount++
			//	actorNagaMap[actor].unMatchCount++
			//	if actor == selectActor {
			//		log.Printf("ダマ推奨 actor: %v, bakaze: %v, kyoku: %v,leftHiNum: %v, reachPredRate: %v", actor, bakaze, kyoku, action.Info.Msg.LeftHaiNum, nextReachRecommend[actor].rate)
			//	}
			//	actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, RECOMMEND_THRETHOLD-nextReachRecommend[actor].rate)
			//	nextReachRecommend[actor].nextShouldReach = false
			//	nextReachRecommend[actor].rate = 0
			//}

			//// 鳴き判断判定ゾーン
			//if nextHuroRecommend[actor].nextShouldHuro && (action.Info.Msg.Type == TYPE_PON || action.Info.Msg.Type == TYPE_CHI) {
			//	actorNagaMap[actor].judgeCount++
			//	actorNagaMap[actor].matchCount++
			//	actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, 0)
			//	nextHuroRecommend[actor].nextShouldHuro = false
			//	nextHuroRecommend[actor].rate = 0
			//	nextHuroRecommend[actor].pattern = ""
			//}
			//if nextHuroRecommend[actor].nextShouldHuro && action.Info.Msg.Type != TYPE_PON && action.Info.Msg.Type != TYPE_CHI {
			//	actorNagaMap[actor].judgeCount++
			//	actorNagaMap[actor].unMatchCount++
			//	actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, nextHuroRecommend[actor].rate)
			//	if nextHuroRecommend[actor].rate > BAD_PLAY_THRETHOLD {
			//		// 悪手を出力してみる
			//		if actor == selectActor {
			//			log.Printf("!!BADPLAYYYYYY 鉄huro!! actor: %v, bakaze: %v, kyoku: %v,leftHiNum: %v, playerChoice: %v, playerChoicePredRate: %v nagaChoice: %v nagaChoicePredRate: %v pattern: %v", actor, bakaze, kyoku, action.Info.Msg.LeftHaiNum, "huroせず", 0, action.Info.Msg.PredDahai, nextHuroRecommend[actor].rate, nextHuroRecommend[actor].pattern)
			//		}
			//		actorNagaMap[actor].BadPlayCount++
			//		nextHuroRecommend[actor].nextShouldHuro = false
			//		nextHuroRecommend[actor].rate = 0
			//		nextHuroRecommend[actor].pattern = ""
			//	} else {
			//		if actor == selectActor {
			//			log.Printf("huro推奨 actor: %v, bakaze: %v, kyoku: %v,leftHiNum: %v, playerChoice: %v, playerChoicePredRate: %v nagaChoice: %v nagaChoicePredRate: %v pattern: %v", actor, bakaze, kyoku, action.Info.Msg.LeftHaiNum, "huroせず", 0, action.Info.Msg.PredDahai, nextHuroRecommend[actor].rate, nextHuroRecommend[actor].pattern)
			//		}
			//		nextHuroRecommend[actor].nextShouldHuro = false
			//		nextHuroRecommend[actor].rate = 0
			//		nextHuroRecommend[actor].pattern = ""
			//	}
			//}
			//if !nextHuroRecommend[actor].nextShouldHuro && (action.Info.Msg.Type == TYPE_PON || action.Info.Msg.Type == TYPE_CHI) {
			//	actorNagaMap[actor].judgeCount++
			//	actorNagaMap[actor].unMatchCount++
			//	if actor == selectActor {
			//		log.Printf("huroしない推奨 actor: %v, bakaze: %v, kyoku: %v,leftHiNum: %v, huroPredRate: %v, huroPredPattern: %v", actor, bakaze, kyoku, action.Info.Msg.LeftHaiNum, nextHuroRecommend[actor].rate, nextHuroRecommend[actor].pattern)
			//	}
			//	actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, RECOMMEND_THRETHOLD-nextHuroRecommend[actor].rate)
			//	nextHuroRecommend[actor].nextShouldHuro = false
			//	nextHuroRecommend[actor].rate = 0
			//	nextHuroRecommend[actor].pattern = ""
			//}

			// 打牌判定ゾーン
			if action.Info.Msg.Type == TYPE_TSUMO || action.Info.Msg.Type == TYPE_PON || action.Info.Msg.Type == TYPE_CHI {
				if action.Reach != 0 {
					nextReachRecommend[actor].rate = action.Reach
					if action.Reach > RECOMMEND_THRETHOLD {
						nextReachRecommend[actor].nextShouldReach = true
					}
				}
				// 手配交換
				// playerTehaiMap.changeTehai(actor, action.Info.Msg.Pai, action.Info.Msg.RealDahai)
				if action.Info.Msg.RealDahai == "" || action.Info.Msg.PredDahai == "" {
					continue
				}
				// NAG推奨打廃できてない場合は、推奨レート次第で加算ポイントを変える
				if action.Info.Msg.RealDahai != action.Info.Msg.PredDahai {
					realDahaiNagaPredRate := action.DahaiPred[realPaiNagaPaiIndexMap[action.Info.Msg.RealDahai]]
					predDahaiNagaPredRate := action.DahaiPred[realPaiNagaPaiIndexMap[action.Info.Msg.PredDahai]]
					if predDahaiNagaPredRate == 0 {
						continue
					}
					actorNagaMap[actor].judgeCount++
					actorNagaMap[actor].unMatchCount++
					if predDahaiNagaPredRate-realDahaiNagaPredRate > BAD_PLAY_THRETHOLD {
						// 悪手を出力してみる
						if actor == selectActor {
							log.Printf("!!BADPLAYYYYYY 打牌選択ミス!! actor: %v, bakaze: %v, kyoku: %v,leftHiNum: %v, playerChoice: %v, playerChoicePredRate: %v nagaChoice: %v nagaChoicePredRate: %v", actor, bakaze, kyoku, action.Info.Msg.LeftHaiNum, action.Info.Msg.RealDahai, realDahaiNagaPredRate, action.Info.Msg.PredDahai, predDahaiNagaPredRate)
						}
						actorNagaMap[actor].BadPlayCount++
						actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, predDahaiNagaPredRate-realDahaiNagaPredRate)
						continue
					}
					if actor == selectActor {
						log.Printf("打牌変更推奨 actor: %v, bakaze: %v, kyoku: %v,leftHiNum: %v, playerChoice: %v, playerChoicePredRate: %v nagaChoice: %v nagaChoicePredRate: %v", actor, bakaze, kyoku, action.Info.Msg.LeftHaiNum, action.Info.Msg.RealDahai, realDahaiNagaPredRate, action.Info.Msg.PredDahai, predDahaiNagaPredRate)
					}
					actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, predDahaiNagaPredRate-realDahaiNagaPredRate)
					continue
				}
				actorNagaMap[actor].matchCount++
				actorNagaMap[actor].minusValueList = append(actorNagaMap[actor].minusValueList, 0)
				actorNagaMap[actor].judgeCount++

			}
			if action.Info.Msg.Type == TYPE_DAHAI {
				if action.Huro != nil {
					for huroAct := 0; huroAct <= 3; huroAct++ {
						huroRate, huroType := getBiggestFloatAndIndex([]float32{action.Huro[huroAct][1], action.Huro[huroAct][2], action.Huro[huroAct][3], action.Huro[huroAct][4], action.Huro[huroAct][5]})
						nextHuroRecommend[huroAct].rate = huroRate
						nextHuroRecommend[huroAct].pattern = strconv.Itoa(huroType)
						if huroRate > RECOMMEND_THRETHOLD {
							nextHuroRecommend[huroAct].nextShouldHuro = true
						}
					}
				}
			}

		}
		kyokuIndex++
	}
	actorNagaMap.culcNagaValue()
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

func getBiggestFloatAndIndex(list []float32) (float32, int) {
	var biggestNumber float32
	var biggestIndex int
	for i, v := range list {
		if v > biggestNumber {
			biggestNumber = v
			biggestIndex = i + 1
		}
	}
	return biggestNumber, biggestIndex
}
