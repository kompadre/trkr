# trkr roadmap

No stable release in sight yet so this is just floats there

- birds view of tracks with phrases (borrow most of the code from track view)
- cycles sizes, tracks logic and phrases logic regardin that
- contextual settings (add basic popup or dialog and start adding specific layouts to cases:
  - [ ] project settings (save, load, bpm...)
  - [ ] track settings (instrument, edit instrument...). one thing fix as a hard design decision is that a track will only have one instrument. That instrument can be either a sample that can be pitched by notes or multisample and the note will determine which sample to play without ever altering its speed
  - [ ] instrument settings
  - [ ] phrase settings (start, end, loop... possibly edit sample)
  - [ ] sample settings (histogram of the waveform and the ability to fade in out and delete selection), maybe apply some baked in effects... 
  - [ ] implement a stream processor, envelopes, effects... 

