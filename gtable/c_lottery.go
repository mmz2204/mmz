package gtable

import "fmt"

/*
@DrawProbability
@csv drawprobability
*/
type DrawProbability struct {
	Id                int
	RewardItem        string
	RewardType        int
	DrawProbability   int
	ProbabilityTimes  int
	ProbabilityChange int
	RealProbability   int //真实概率
	FirstRand         int
	Limit1            int
	Limit2            int
}

func (pData *DrawProbability) OnInit() {
	fmt.Println(pData.Id)
}
