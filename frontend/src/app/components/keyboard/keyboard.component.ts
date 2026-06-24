import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface KeyInfo {
  code: string;
  label: string;
  shiftLabel?: string;
  altGrLabel?: string;
  finger: string;
  hand: 'left' | 'right' | 'thumb';
  widthClass?: string;
}

@Component({
  selector: 'app-keyboard',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './keyboard.component.html',
  styleUrls: ['./keyboard.component.css']
})
export class KeyboardComponent {
  @Input() activeKey: string = '';       // Key to highlight (e.g. 'q', 'KeyA', 'Quote')
  @Input() activeFinger: string = '';    // Finger to highlight (e.g. 'left-pinky', 'right-index')
  @Input() shiftActive: boolean = false; // Is Shift required for this character?
  @Input() altGrActive: boolean = false; // Is AltGr required for this character?

  // Mac ISO US-International layout
  readonly row1: KeyInfo[] = [
    { code: 'Backquote', label: '§', shiftLabel: '±', finger: 'left-pinky', hand: 'left' },
    { code: 'Digit1', label: '1', shiftLabel: '!', finger: 'left-pinky', hand: 'left' },
    { code: 'Digit2', label: '2', shiftLabel: '@', finger: 'left-ring', hand: 'left' },
    { code: 'Digit3', label: '3', shiftLabel: '#', finger: 'left-middle', hand: 'left' },
    { code: 'Digit4', label: '4', shiftLabel: '$', finger: 'left-index', hand: 'left' },
    { code: 'Digit5', label: '5', shiftLabel: '%', finger: 'left-index', hand: 'left' },
    { code: 'Digit6', label: '6', shiftLabel: '^', finger: 'right-index', hand: 'right' },
    { code: 'Digit7', label: '7', shiftLabel: '&', finger: 'right-index', hand: 'right' },
    { code: 'Digit8', label: '8', shiftLabel: '*', finger: 'right-middle', hand: 'right' },
    { code: 'Digit9', label: '9', shiftLabel: '(', finger: 'right-ring', hand: 'right' },
    { code: 'Digit0', label: '0', shiftLabel: ')', finger: 'right-pinky', hand: 'right' },
    { code: 'Minus', label: '-', shiftLabel: '_', finger: 'right-pinky', hand: 'right' },
    { code: 'Equal', label: '=', shiftLabel: '+', finger: 'right-pinky', hand: 'right' },
    { code: 'Backspace', label: 'Delete', finger: 'right-pinky', hand: 'right', widthClass: 'w-backspace' }
  ];

  readonly row2: KeyInfo[] = [
    { code: 'Tab', label: 'Tab', finger: 'left-pinky', hand: 'left', widthClass: 'w-tab' },
    { code: 'KeyQ', label: 'q', shiftLabel: 'Q', finger: 'left-pinky', hand: 'left' },
    { code: 'KeyW', label: 'w', shiftLabel: 'W', finger: 'left-ring', hand: 'left' },
    { code: 'KeyE', label: 'e', shiftLabel: 'E', finger: 'left-middle', hand: 'left' },
    { code: 'KeyR', label: 'r', shiftLabel: 'R', finger: 'left-index', hand: 'left' },
    { code: 'KeyT', label: 't', shiftLabel: 'T', finger: 'left-index', hand: 'left' },
    { code: 'KeyY', label: 'y', shiftLabel: 'Y', finger: 'right-index', hand: 'right' },
    { code: 'KeyU', label: 'u', shiftLabel: 'U', finger: 'right-index', hand: 'right' },
    { code: 'KeyI', label: 'i', shiftLabel: 'I', finger: 'right-middle', hand: 'right' },
    { code: 'KeyO', label: 'o', shiftLabel: 'O', finger: 'right-ring', hand: 'right' },
    { code: 'KeyP', label: 'p', shiftLabel: 'P', finger: 'right-pinky', hand: 'right' },
    { code: 'BracketLeft', label: '[', shiftLabel: '{', finger: 'right-pinky', hand: 'right' },
    { code: 'BracketRight', label: ']', shiftLabel: '}', finger: 'right-pinky', hand: 'right' },
    { code: 'Enter', label: 'Return', finger: 'right-pinky', hand: 'right', widthClass: 'w-enter' }
  ];

  readonly row3: KeyInfo[] = [
    { code: 'CapsLock', label: 'Caps Lock', finger: 'left-pinky', hand: 'left', widthClass: 'w-caps' },
    { code: 'KeyA', label: 'a', shiftLabel: 'A', finger: 'left-pinky', hand: 'left' },
    { code: 'KeyS', label: 's', shiftLabel: 'S', finger: 'left-ring', hand: 'left' },
    { code: 'KeyD', label: 'd', shiftLabel: 'D', finger: 'left-middle', hand: 'left' },
    { code: 'KeyF', label: 'f', shiftLabel: 'F', finger: 'left-index', hand: 'left' },
    { code: 'KeyG', label: 'g', shiftLabel: 'G', finger: 'left-index', hand: 'left' },
    { code: 'KeyH', label: 'h', shiftLabel: 'H', finger: 'right-index', hand: 'right' },
    { code: 'KeyJ', label: 'j', shiftLabel: 'J', finger: 'right-index', hand: 'right' },
    { code: 'KeyK', label: 'k', shiftLabel: 'K', finger: 'right-middle', hand: 'right' },
    { code: 'KeyL', label: 'l', shiftLabel: 'L', finger: 'right-ring', hand: 'right' },
    { code: 'Semicolon', label: ';', shiftLabel: ':', finger: 'right-pinky', hand: 'right' },
    { code: 'Quote', label: '\'', shiftLabel: '"', finger: 'right-pinky', hand: 'right' },
    { code: 'Backslash', label: '\\', shiftLabel: '|', finger: 'right-pinky', hand: 'right' }
  ];

  readonly row4: KeyInfo[] = [
    { code: 'ShiftLeft', label: 'Shift', finger: 'left-pinky', hand: 'left', widthClass: 'w-shift-left' },
    { code: 'IntlBackslash', label: '`', shiftLabel: '~', finger: 'left-pinky', hand: 'left' }, // Mac ISO specific
    { code: 'KeyZ', label: 'z', shiftLabel: 'Z', finger: 'left-pinky', hand: 'left' },
    { code: 'KeyX', label: 'x', shiftLabel: 'X', finger: 'left-ring', hand: 'left' },
    { code: 'KeyC', label: 'c', shiftLabel: 'C', finger: 'left-middle', hand: 'left' },
    { code: 'KeyV', label: 'v', shiftLabel: 'V', finger: 'left-index', hand: 'left' },
    { code: 'KeyB', label: 'b', shiftLabel: 'B', finger: 'left-index', hand: 'left' },
    { code: 'KeyN', label: 'n', shiftLabel: 'N', finger: 'right-index', hand: 'right' },
    { code: 'KeyM', label: 'm', shiftLabel: 'M', finger: 'right-index', hand: 'right' },
    { code: 'Comma', label: ',', shiftLabel: '<', finger: 'right-middle', hand: 'right' },
    { code: 'Period', label: '.', shiftLabel: '>', finger: 'right-ring', hand: 'right' },
    { code: 'Slash', label: '/', shiftLabel: '?', finger: 'right-pinky', hand: 'right' },
    { code: 'ShiftRight', label: 'Shift', finger: 'right-pinky', hand: 'right', widthClass: 'w-shift-right' }
  ];

  readonly row5: KeyInfo[] = [
    { code: 'ControlLeft', label: '⌃', finger: 'left-pinky', hand: 'left', widthClass: 'w-ctrl' },
    { code: 'AltLeft', label: '⌥', finger: 'left-ring', hand: 'left', widthClass: 'w-option' },
    { code: 'MetaLeft', label: '⌘', finger: 'left-thumb', hand: 'left', widthClass: 'w-cmd' },
    { code: 'Space', label: '', finger: 'thumb', hand: 'thumb', widthClass: 'w-space' },
    { code: 'MetaRight', label: '⌘', finger: 'right-thumb', hand: 'right', widthClass: 'w-cmd' },
    { code: 'AltRight', label: '⌥ GR', finger: 'right-ring', hand: 'right', widthClass: 'w-option' }, // AltGr for international characters
    { code: 'ControlRight', label: '⌃', finger: 'right-pinky', hand: 'right', widthClass: 'w-ctrl' }
  ];

  isKeyHighlighted(keyInfo: KeyInfo): boolean {
    // Matches by key code or exact key symbol representation
    return this.activeKey === keyInfo.code || 
           this.activeKey.toLowerCase() === keyInfo.label.toLowerCase() ||
           !!(keyInfo.shiftLabel && this.activeKey.toLowerCase() === keyInfo.shiftLabel.toLowerCase());
  }

  isModifierHighlighted(code: string): boolean {
    if (code === 'ShiftLeft' || code === 'ShiftRight') {
      return this.shiftActive;
    }
    if (code === 'AltRight') {
      return this.altGrActive;
    }
    return false;
  }
}
