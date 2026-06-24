import { Injectable, signal } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class SoundService {
  readonly soundEnabled = signal<boolean>(true);
  readonly switchType = signal<'blue' | 'brown' | 'typewriter'>('blue');

  private audioCtx: AudioContext | null = null;

  private initAudio() {
    if (!this.audioCtx) {
      this.audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)();
    }
    if (this.audioCtx.state === 'suspended') {
      this.audioCtx.resume();
    }
  }

  playKeySound() {
    if (!this.soundEnabled()) return;
    this.initAudio();
    if (!this.audioCtx) return;

    const ctx = this.audioCtx;
    const now = ctx.currentTime;

    // Create click noise node (for tactile transient)
    const bufferSize = ctx.sampleRate * 0.02; // 20ms
    const buffer = ctx.createBuffer(1, bufferSize, ctx.sampleRate);
    const data = buffer.getChannelData(0);
    for (let i = 0; i < bufferSize; i++) {
      data[i] = Math.random() * 2 - 1;
    }

    const noise = ctx.createBufferSource();
    noise.buffer = buffer;

    const noiseFilter = ctx.createBiquadFilter();
    const noiseGain = ctx.createGain();

    // Resonator (for keycap/housing acoustics)
    const osc = ctx.createOscillator();
    const oscGain = ctx.createGain();

    noise.connect(noiseFilter);
    noiseFilter.connect(noiseGain);
    noiseGain.connect(ctx.destination);

    osc.connect(oscGain);
    oscGain.connect(ctx.destination);

    const type = this.switchType();

    if (type === 'blue') {
      // Sharp click + lower clack
      noiseFilter.type = 'bandpass';
      noiseFilter.frequency.setValueAtTime(3500, now);
      noiseFilter.Q.setValueAtTime(3, now);

      noiseGain.gain.setValueAtTime(0.08, now);
      noiseGain.gain.exponentialRampToValueAtTime(0.001, now + 0.005);

      osc.type = 'sine';
      osc.frequency.setValueAtTime(250, now);
      oscGain.gain.setValueAtTime(0.12, now);
      oscGain.gain.exponentialRampToValueAtTime(0.001, now + 0.03);

      noise.start(now);
      osc.start(now);
      noise.stop(now + 0.01);
      osc.stop(now + 0.04);

    } else if (type === 'brown') {
      // Dull clack
      noiseFilter.type = 'lowpass';
      noiseFilter.frequency.setValueAtTime(1000, now);

      noiseGain.gain.setValueAtTime(0.05, now);
      noiseGain.gain.exponentialRampToValueAtTime(0.001, now + 0.008);

      osc.type = 'triangle';
      osc.frequency.setValueAtTime(180, now);
      oscGain.gain.setValueAtTime(0.08, now);
      oscGain.gain.exponentialRampToValueAtTime(0.001, now + 0.04);

      noise.start(now);
      osc.start(now);
      noise.stop(now + 0.01);
      osc.stop(now + 0.05);

    } else if (type === 'typewriter') {
      // Metallic snap + longer resonance
      noiseFilter.type = 'bandpass';
      noiseFilter.frequency.setValueAtTime(2800, now);
      noiseFilter.Q.setValueAtTime(5, now);

      noiseGain.gain.setValueAtTime(0.1, now);
      noiseGain.gain.exponentialRampToValueAtTime(0.001, now + 0.008);

      osc.type = 'sawtooth';
      osc.frequency.setValueAtTime(500, now);
      // Fast pitch drop to simulate typewriter lever impact
      osc.frequency.exponentialRampToValueAtTime(80, now + 0.05);

      oscGain.gain.setValueAtTime(0.06, now);
      oscGain.gain.exponentialRampToValueAtTime(0.001, now + 0.08);

      noise.start(now);
      osc.start(now);
      noise.stop(now + 0.01);
      osc.stop(now + 0.09);
    }
  }

  playErrorSound() {
    if (!this.soundEnabled()) return;
    this.initAudio();
    if (!this.audioCtx) return;

    const ctx = this.audioCtx;
    const now = ctx.currentTime;

    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.type = 'square';
    osc.frequency.setValueAtTime(130, now); // Low buzz
    osc.frequency.linearRampToValueAtTime(100, now + 0.15);

    gain.gain.setValueAtTime(0.15, now);
    gain.gain.linearRampToValueAtTime(0.001, now + 0.15);

    osc.start(now);
    osc.stop(now + 0.16);
  }

  playSuccessSound() {
    if (!this.soundEnabled()) return;
    this.initAudio();
    if (!this.audioCtx) return;

    const ctx = this.audioCtx;
    const now = ctx.currentTime;

    // Arpeggio notes: C4, E4, G4, C5
    const notes = [261.63, 329.63, 392.00, 523.25];
    const duration = 0.1;
    const delay = 0.08;

    notes.forEach((freq, idx) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'sine';
      osc.frequency.setValueAtTime(freq, now + idx * delay);

      gain.gain.setValueAtTime(0, now + idx * delay);
      gain.gain.linearRampToValueAtTime(0.12, now + idx * delay + 0.01);
      gain.gain.exponentialRampToValueAtTime(0.001, now + idx * delay + duration);

      osc.start(now + idx * delay);
      osc.stop(now + idx * delay + duration + 0.01);
    });
  }
}
