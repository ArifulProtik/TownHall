import { createSlice, type PayloadAction } from '@reduxjs/toolkit';
import type { RootState } from '@/app/store';

interface AuthState {
  accessToken: string | null;
}

const initialState: AuthState = {
  accessToken: null,
};

export const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (state, action: PayloadAction<{ accessToken: string }>) => {
      state.accessToken = action.payload.accessToken;
    },
    clearCredentials: (state) => {
      state.accessToken = null;
    },
  },
});

export const { setCredentials, clearCredentials } = authSlice.actions;
export default authSlice.reducer;

export const selectAccessToken = (state: RootState): string | null => state.auth.accessToken;
export const selectIsAuthenticated = (state: RootState): boolean =>
  state.auth.accessToken !== null;
