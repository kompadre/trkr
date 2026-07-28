package view

import (
	"trkr"
	ev "trkr/internal/events"
	"trkr/internal/ui"
	"trkr/internal/ui/widget"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var volKnob *widget.Knob
var typeKnob *widget.Knob
var slotKnob *widget.Knob
var freqKnob *widget.Knob
var modAmtKnob *widget.Knob
var modRatioKnob *widget.Knob
var ampDecayKnob *widget.Knob
var pitchDecayKnob *widget.Knob
var noiseDecayKnob *widget.Knob
var noiseMixKnob *widget.Knob

var currentPercSlot int

func CreateInstrumentSettings(parent *ui.Element) {
	core := ui.NewElementCoreInstance(showInstrumentSettings, hideInstrumentSettings, instrumentHandleInputs, drawInstrumentSettings)
	uiElem := ui.NewElement(0, 0, int32(ui.GetOptions().ScreenWidth), 300, core, parent)
	uiElem.Name = "Instrument Settings"
	uiElem.Visible = false
	uiElem.IsAnchor = true
	uiElem.TopPadding = 40
	uiElem.LeftPadding = 10
	ui.InstrumentDialog = uiElem

	row1 := ui.NewRow(80, uiElem)

	// Volume Knob
	volKnob = widget.NewKnob("Volume", [4]int{6, 6, 6, 6}, 0, 100, 100, nil, row1)
	volKnob.OnChange = func(val float64) {
		currentTrack := CurrentTrack()
		currentTrack.Volume = val / 100.0
	}

	// Sample Source Type Knob
	typeKnob = widget.NewKnob("Type", [4]int{6, 6, 6, 6}, 0, float64(trkr.SampleSourceTypePerc), 0, nil, row1)
	typeKnob.OnChange = func(val float64) {
		currentTrack := CurrentTrack()
		currentTrack.Instrument.SampleSourceType = trkr.SampleSourceType(int(val))
	}

	// Percussion Rows (only visible when type is Perc)
	percRow1 := ui.NewRow(80, uiElem)
	slotKnob = widget.NewKnob("Slot", [4]int{3, 3, 3, 3}, 0, 3, 0, nil, percRow1)
	slotKnob.OnChange = func(val float64) {
		currentPercSlot = int(val)
		syncInstrumentSettings()
	}

	freqKnob = widget.NewKnob("Freq", [4]int{3, 3, 3, 3}, 20, 2000, 440, nil, percRow1)
	freqKnob.OnChange = func(val float64) {
		CurrentTrack().Instrument.Percs[currentPercSlot].Freq = val
	}

	modAmtKnob = widget.NewKnob("ModAmt", [4]int{3, 3, 3, 3}, 0, 20, 0, nil, percRow1)
	modAmtKnob.OnChange = func(val float64) {
		CurrentTrack().Instrument.Percs[currentPercSlot].ModAmt = val
	}

	modRatioKnob = widget.NewKnob("ModRat", [4]int{3, 3, 3, 3}, 0, 10, 0, nil, percRow1)
	modRatioKnob.OnChange = func(val float64) {
		CurrentTrack().Instrument.Percs[currentPercSlot].ModRatio = val
	}

	percRow2 := ui.NewRow(80, uiElem)
	ampDecayKnob = widget.NewKnob("AmpDec", [4]int{3, 3, 3, 3}, 0.0, 1.0, 0.99, nil, percRow2)
	ampDecayKnob.Step = 0.001
	ampDecayKnob.OnChange = func(val float64) {
		CurrentTrack().Instrument.Percs[currentPercSlot].AmpDecay = val
	}

	pitchDecayKnob = widget.NewKnob("PitDec", [4]int{3, 3, 3, 3}, 0.0, 1.0, 0.99, nil, percRow2)
	pitchDecayKnob.Step = 0.001
	pitchDecayKnob.OnChange = func(val float64) {
		CurrentTrack().Instrument.Percs[currentPercSlot].PitchDecay = val
	}

	noiseDecayKnob = widget.NewKnob("NoiDec", [4]int{3, 3, 3, 3}, 0.0, 1.0, 0.99, nil, percRow2)
	noiseDecayKnob.Step = 0.001
	noiseDecayKnob.OnChange = func(val float64) {
		CurrentTrack().Instrument.Percs[currentPercSlot].NoiseDecay = val
	}

	noiseMixKnob = widget.NewKnob("NoiMix", [4]int{3, 3, 3, 3}, 0, 1, 0, nil, percRow2)
	noiseMixKnob.Step = 0.01
	noiseMixKnob.OnChange = func(val float64) {
		CurrentTrack().Instrument.Percs[currentPercSlot].NoiseMix = val
	}

	row2 := ui.NewRow(40, uiElem)
	buttonCol := [4]int{3, 3, 3, 3}
	widget.NewButton("PATCHES", buttonCol, 30, func(_ any) bool {
		showPatchBrowser()
		return true
	}, row2)
	widget.NewButton("BACK", buttonCol, 30, func(_ any) bool {
		hideInstrumentSettings()
		return true
	}, row2)
}

func hideInstrumentSettings() {
	ui.InstrumentDialog.Visible = false
	ui.TrackDialog.Visible = true
	ui.InstrumentDialog.Parent.SetFocus(ui.TrackDialog)
}

func showInstrumentSettings() {
	if ui.InstrumentDialog.Visible == false {
		ui.InstrumentDialog.Visible = true
		syncInstrumentSettings()
		ui.InstrumentDialog.Parent.SetFocus(ui.InstrumentDialog)
	}
}

func syncInstrumentSettings() {
	track := CurrentTrack()
	if track == nil {
		return
	}
	volKnob.Value = track.Volume * 100
	if track.Instrument != nil {
		typeKnob.Value = float64(track.Instrument.SampleSourceType)

		isPerc := track.Instrument.SampleSourceType == trkr.SampleSourceTypePerc

		if isPerc {
			p := &track.Instrument.Percs[currentPercSlot]
			slotKnob.Value = float64(currentPercSlot)
			freqKnob.Value = p.Freq
			modAmtKnob.Value = p.ModAmt
			modRatioKnob.Value = p.ModRatio
			ampDecayKnob.Value = p.AmpDecay
			pitchDecayKnob.Value = p.PitchDecay
			noiseDecayKnob.Value = p.NoiseDecay
			noiseMixKnob.Value = p.NoiseMix
		}
	}
}

func drawInstrumentSettings(ctx ev.EventContext, hasFocus bool, isHighlighted bool) bool {
	// ... (rest of draw logic remains same)
	rec := rl.Rectangle{
		X:      0,
		Y:      0,
		Width:  float32(rl.GetScreenWidth()),
		Height: float32(rl.GetScreenHeight()),
	}
	rl.DrawRectangleRec(rec, ui.WindowBg1)
	ui.DrawText("Instrument Settings", 10, 10, 20, ui.WindowFg1)

	track := CurrentTrack()
	if track != nil && track.Instrument != nil {
		isPerc := track.Instrument.SampleSourceType == trkr.SampleSourceTypePerc
		ui.InstrumentDialog.Children[1].Visible = isPerc
		ui.InstrumentDialog.Children[2].Visible = isPerc
	}

	return false
}

func instrumentHandleInputs(input *ev.InputSnapshot, el *ui.Element) bool {
	// Root level handle input for the dialog.
	// Since navigation is now handled by the UI system automatically,
	// we only need to handle global dialog actions here.
	if input.Tick(ev.InputKindB) == 1 {
		hideInstrumentSettings()
		return true
	}
	return false
}

