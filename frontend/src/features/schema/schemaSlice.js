import { createSlice } from '@reduxjs/toolkit';

const initialState = {
    loading: false,
    error: '',
    value: {
        version: '1',
        schema: [],
    },
    generated: {},
};

export const schemaSlice = createSlice({
    name: 'schema',
    initialState,
    reducers: {
        update: (state, action) => {
            state.value = action.payload;
            state.generated = {};
            state.loading = false;
            state.error = '';
        },
        updateGenerated: (state, action) => {
            const { id, value } = action.payload;

            state.generated[id] = value;
            state.loading = false;
            state.error = '';
        },
        loading: (state, action) => {
            state.loading = action.payload;
        },
        error: (state, action) => {
            state.error = action.payload;
        },
    },
});

export const { update, updateGenerated, loading, error } = schemaSlice.actions;
export default schemaSlice.reducer;
