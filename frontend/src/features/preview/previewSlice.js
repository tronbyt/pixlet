import { createSlice } from '@reduxjs/toolkit';

export const previewSlice = createSlice({
    name: 'preview',
    initialState: {
        loading: false,
        value: {
            img: '',
            img_type: '',
            title: 'Pixlet',
            width: 64,
            height: 32,
            is2x: null,
            timezone: '',
            locale: '',
        }
    },
    reducers: {
        update: (state = initialState, action) => {
            let up = state;

            if ('img' in action.payload) {
                up.value.img = action.payload.img;
            }

            if ('img_type' in action.payload) {
                up.value.img_type = action.payload.img_type;
            }

            if ('title' in action.payload) {
                up.value.title = action.payload.title;
            }

            if ('width' in action.payload) {
                up.value.width = action.payload.width;
            }

            if ('height' in action.payload) {
                up.value.height = action.payload.height;
            }

            if ('is2x' in action.payload) {
                up.value.is2x = action.payload.is2x;
            }

            if ('timezone' in action.payload) {
                up.value.timezone = action.payload.timezone || '';
            }

            if ('locale' in action.payload) {
                up.value.locale = action.payload.locale || '';
            }

            return up;
        },
        loading: (state = initialState, action) => {
            return { ...state, loading: action.payload }
        },
        setScale: (state = initialState, action) => {
            return {
                ...state,
                value: {
                    ...state.value,
                    is2x: action.payload,
                }
            }
        },
        setTimezone: (state = initialState, action) => {
            return {
                ...state,
                value: {
                    ...state.value,
                    timezone: action.payload || '',
                }
            }
        },
        setLocale: (state = initialState, action) => {
            return {
                ...state,
                value: {
                    ...state.value,
                    locale: action.payload || '',
                }
            }
        },
    },
});

export const { update, loading, setScale, setTimezone, setLocale } = previewSlice.actions;
export default previewSlice.reducer;
