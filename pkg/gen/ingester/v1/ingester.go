package ingesterv1

func (x *MergeProfilesStacktracesResponse) SetSelectedProfiles(p *ProfileSets) {
	if x != nil {
		x.SelectedProfiles = p
	}
}

func (x *MergeProfilesLabelsResponse) SetSelectedProfiles(p *ProfileSets) {
	if x != nil {
		x.SelectedProfiles = p
	}
}
