package vpn

func (u UserStatsDatasets) Len() int {
	return len(u)
}

func (u UserStatsDatasets) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u UserStatsDatasets) Less(i, j int) bool {
	return u[i].Label < u[j].Label
}
