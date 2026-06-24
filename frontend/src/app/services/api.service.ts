import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Profile {
  id?: number;
  name: string;
  xp?: number;
  level?: number;
  streak?: number;
  theme?: string;
  sound_enabled?: boolean;
  switch_type?: string;
  last_active?: string;
  created_at?: string;
}

export interface Exercise {
  id: string;
  title: string;
  content: string;
  category: string;
  difficulty: string;
  is_endurance: boolean;
  unlocked?: boolean;
  high_score_wpm?: number;
  high_score_accuracy?: number;
  arcade_wpm?: number;
  arcade_accuracy?: number;
  play_count?: number;
}

export interface KeyProgress {
  attempts: number;
  errors: number;
  latency_ms: number;
}

export interface SaveSessionPayload {
  profile_id: number;
  exercise_id: string;
  mode: string; // 'lesson', 'arcade', 'endurance', 'retry'
  wpm: number;
  accuracy: number;
  duration_seconds: number;
  raw_data_json: string;
  backspaces_used: number;
  keys: { [key: string]: KeyProgress };
}

export interface SaveSessionResponse {
  xp_gained: number;
  new_xp: number;
  new_level: number;
  new_streak: number;
  level_up: boolean;
  new_achievements: string[];
}

export interface KeyStats {
  char: string;
  error_rate: number;
}

export interface SessionWPM {
  completed_at: string;
  wpm: number;
}

export interface SummaryStats {
  total_sessions: number;
  average_wpm: number;
  average_accuracy: number;
  total_duration: number;
  weakest_keys: KeyStats[];
  progress_wpm: SessionWPM[];
  achievements: string[];
}

export interface RetryItem {
  exercise_id: string;
  title: string;
  category: string;
  difficulty: string;
  last_wpm: number;
  last_accuracy: number;
}

export interface AchievementItem {
  code: string;
  unlocked_at: string;
}

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  private http = inject(HttpClient);
  private baseUrl = '/api'; // Handled by Nginx proxy pass

  getProfiles(): Observable<Profile[]> {
    return this.http.get<Profile[]>(`${this.baseUrl}/profiles`);
  }

  createProfile(profile: Profile): Observable<Profile> {
    return this.http.post<Profile>(`${this.baseUrl}/profiles`, profile);
  }

  getProfile(id: number): Observable<Profile> {
    return this.http.get<Profile>(`${this.baseUrl}/profiles/${id}`);
  }

  updateProfileSettings(profileId: number, settings: { theme: string, sound_enabled: boolean, switch_type: string }): Observable<{ result: string }> {
    return this.http.put<{ result: string }>(`${this.baseUrl}/profiles/${profileId}/settings`, settings);
  }

  deleteProfile(id: number): Observable<{ result: string }> {
    return this.http.delete<{ result: string }>(`${this.baseUrl}/profiles/${id}`);
  }

  getExercises(category?: string, difficulty?: string, isEndurance?: boolean, profileId?: number): Observable<Exercise[]> {
    let params: any = {};
    if (category) params.category = category;
    if (difficulty) params.difficulty = difficulty;
    if (isEndurance !== undefined) params.is_endurance = isEndurance ? 'true' : 'false';
    if (profileId) params.profile_id = profileId.toString();

    return this.http.get<Exercise[]>(`${this.baseUrl}/exercises`, { params });
  }

  saveSession(payload: SaveSessionPayload): Observable<SaveSessionResponse> {
    return this.http.post<SaveSessionResponse>(`${this.baseUrl}/sessions`, payload);
  }

  getStats(profileId: number): Observable<SummaryStats> {
    return this.http.get<SummaryStats>(`${this.baseUrl}/profiles/${profileId}/stats`);
  }

  getRetries(profileId: number): Observable<RetryItem[]> {
    return this.http.get<RetryItem[]>(`${this.baseUrl}/profiles/${profileId}/retries`);
  }

  getAchievements(profileId: number): Observable<AchievementItem[]> {
    return this.http.get<AchievementItem[]>(`${this.baseUrl}/profiles/${profileId}/achievements`);
  }
}
