package intermodal

// Move plans (rdt: value_objects.py MovePlan / LiveLoad / LiveUnload) describe a
// kind of drayage move and know how to expand themselves into an ordered agenda
// spec. They are pure value objects — no aggregate state — so the agenda for a
// given plan is deterministic and unit-testable on its own. DispatchFactory
// (ported later with the aggregate) turns a plan's spec into a queued Dispatch.

// MovePlan is a planned move that can produce the ordered agenda a Dispatch
// will execute (rdt: MovePlan protocol).
type MovePlan interface {
	PlanType() PlanType
	ToAgendaSpec() []AgendaItemSpec
}

// facilityAt returns a pointer to a copy of f, for the optional Facility field.
func facilityAt(f FacilityID) *FacilityID { return &f }

// containerDetails returns item details carrying a copy of c.
func containerDetails(c ContainerNumber) *AgendaItemDetails {
	return &AgendaItemDetails{Container: &c}
}

// LiveLoad: pick up an empty container, deliver it to a dock door to be loaded,
// wait, take the now-loaded container, and drop it at the destination ramp
// (rdt: LiveLoad). The premounted variant skips the separate chassis
// acquire/return because the container arrives already mounted on a chassis.
type LiveLoad struct {
	ContainerNumber      ContainerNumber
	ContainerPickupRamp  FacilityID
	ContainerDropoffRamp FacilityID
	DockDoor             FacilityID
	Premounted           bool
}

// PlanType identifies this as a live-load move.
func (p LiveLoad) PlanType() PlanType { return PlanLiveLoad }

// ToAgendaSpec expands the plan into its ordered agenda: set-up, then the core
// deliver/hold/take at the dock, then wrap-up. Order and actions mirror rdt
// exactly so the two apps plan identical agendas.
func (p LiveLoad) ToAgendaSpec() []AgendaItemSpec {
	setUp := []AgendaItemSpec{
		{Action: ActionAcquireChassis},
		{
			Action:   ActionPickUpEmptyContainer,
			Facility: facilityAt(p.ContainerPickupRamp),
			Details:  containerDetails(p.ContainerNumber),
		},
	}
	core := []AgendaItemSpec{
		{Action: ActionDeliverEmptyContainer, Facility: facilityAt(p.DockDoor)},
		{Action: ActionHoldForLoading, Facility: facilityAt(p.DockDoor)},
		{Action: ActionTakeLoadedContainer, Facility: facilityAt(p.DockDoor)},
	}
	wrapUp := []AgendaItemSpec{
		{Action: ActionDropLoadedContainer, Facility: facilityAt(p.ContainerDropoffRamp)},
		{Action: ActionReturnChassis},
	}
	if p.Premounted {
		setUp = []AgendaItemSpec{
			{
				Action:   ActionPickUpEmptyPremountedContainer,
				Facility: facilityAt(p.ContainerPickupRamp),
				Details:  containerDetails(p.ContainerNumber),
			},
		}
		wrapUp = []AgendaItemSpec{
			{Action: ActionDropLoadedPremountedContainer, Facility: facilityAt(p.ContainerDropoffRamp)},
		}
	}
	return concat(setUp, core, wrapUp)
}

// LiveUnload: pick up a loaded container, deliver it to a dock door to be
// emptied, wait, take the now-empty container, and drop it at the destination
// ramp (rdt: LiveUnload). The premounted variant skips the separate chassis
// acquire/return.
type LiveUnload struct {
	ContainerNumber      ContainerNumber
	ContainerPickupRamp  FacilityID
	ContainerDropoffRamp FacilityID
	DockDoor             FacilityID
	Premounted           bool
}

// PlanType identifies this as a live-unload move.
func (p LiveUnload) PlanType() PlanType { return PlanLiveUnload }

// ToAgendaSpec expands the plan into its ordered agenda, mirroring rdt.
func (p LiveUnload) ToAgendaSpec() []AgendaItemSpec {
	setUp := []AgendaItemSpec{
		{Action: ActionAcquireChassis},
		{
			Action:   ActionPickUpLoadedContainer,
			Facility: facilityAt(p.ContainerPickupRamp),
			Details:  containerDetails(p.ContainerNumber),
		},
	}
	core := []AgendaItemSpec{
		{Action: ActionDeliverLoadedContainer, Facility: facilityAt(p.DockDoor)},
		{Action: ActionHoldForEmptying, Facility: facilityAt(p.DockDoor)},
		{Action: ActionTakeEmptyContainer, Facility: facilityAt(p.DockDoor)},
	}
	wrapUp := []AgendaItemSpec{
		{Action: ActionDropEmptyContainer, Facility: facilityAt(p.ContainerDropoffRamp)},
		{Action: ActionReturnChassis},
	}
	if p.Premounted {
		setUp = []AgendaItemSpec{
			{
				Action:   ActionPickUpLoadedPremountedContainer,
				Facility: facilityAt(p.ContainerPickupRamp),
				Details:  containerDetails(p.ContainerNumber),
			},
		}
		wrapUp = []AgendaItemSpec{
			{Action: ActionDropEmptyPremountedContainer, Facility: facilityAt(p.ContainerDropoffRamp)},
		}
	}
	return concat(setUp, core, wrapUp)
}

// concat flattens agenda segments into a single ordered slice.
func concat(segments ...[]AgendaItemSpec) []AgendaItemSpec {
	var n int
	for _, s := range segments {
		n += len(s)
	}
	out := make([]AgendaItemSpec, 0, n)
	for _, s := range segments {
		out = append(out, s...)
	}
	return out
}
