import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import dayjs from 'dayjs';

import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';
import TextField from '@mui/material/TextField';

import { set, remove } from '../../config/configSlice'


export default function DateTime({ field }) {
    const [dateTime, setDateTime] = useState(dayjs());
    const config = useSelector(state => state.config);
    const dispatch = useDispatch();

    useEffect(() => {
        if (field.id in config) {
            setDateTime(dayjs(config[field.id].value));
        }
    }, [config]);

    const onChange = (timestamp) => {
        if (!timestamp) {
            setDateTime(dayjs());
            dispatch(remove(field.id));
            return;
        }

        setDateTime(timestamp);
        dispatch(set({
            id: field.id,
            value: timestamp.toISOString(),
        }));
    }

    return (
        <LocalizationProvider dateAdapter={AdapterDayjs}>
            <DateTimePicker
                renderInput={(props) => <TextField {...props} />}
                label={field.name}
                value={dateTime}
                onChange={onChange}
                onError={console.log}
            />
        </LocalizationProvider>
    );
}