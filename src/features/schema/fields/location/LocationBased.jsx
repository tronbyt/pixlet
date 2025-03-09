import { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';

import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import InputSlider from './InputSlider';
import { set } from '../../../config/configSlice';
import { callHandler } from '../../../handlers/actions';

export default function LocationBased({ field }) {
    const [value, setValue] = useState({
        // Default to Brooklyn, because that's where tidbyt folks
        // are and  we can only dispatch a location object which
        // has all fields set.
        'lat': 40.678,
        'lng': -73.944,
        'locality': 'Brooklyn, New York',
        'timezone': 'America/New_York',
        // But overwrite with app-specific defaults set in config.
        'display': '',
        'value': '',
        ...field.default
    });

    const config = useSelector(state => state.config);
    const dispatch = useDispatch();
    const handlerResults = useSelector(state => state.handlers)

    const getLocationAsJson = (v) => {
        return JSON.stringify({
            lat: v.lat,
            lng: v.lng,
            locality: v.locality,
            timezone: v.timezone
        });
    }

    useEffect(() => {
        if (field.id in config) {
            let v = JSON.parse(config[field.id].value);
            setValue(v);
            callHandler(field.id, field.handler, getLocationAsJson(v));
        } else if (field.default) {
            value = field.default;
            dispatch(set({
                id: field.id,
                value: JSON.stringify(field.default),
            }));
        }
    }, [config]);

    useEffect(() => {
        if (!(field.id in config)) {
            callHandler(field.id, field.handler, getLocationAsJson(value));
        }
    }, []);

    const setLocationPart = (partName, partValue) => {
        let newValue = { ...value };
        newValue[partName] = partValue;
        setValue(newValue);
        dispatch(set({
            id: field.id,
            value: JSON.stringify(newValue),
        }));
        callHandler(field.id, field.handler, getLocationAsJson(newValue));
    }

    const truncateLatLng = (value) => {
        return String(Number(value).toFixed(3));
    }

    const onChangeLatitude = (event) => {
        setLocationPart('lat', truncateLatLng(event.target.value));
    }

    const onChangeLongitude = (event) => {
        setLocationPart('lng', truncateLatLng(event.target.value));
    }

    const onChangeLocality = (event) => {
        setLocationPart('locality', event.target.value);
    }

    const onChangeTimezone = (event) => {
        setLocationPart('timezone', event.target.value);
    }

    const onChangeOption = (event) => {
        let newValue = { ...value };
        newValue.value = event.target.value;
        const selectedOption = options.find(option => option.value === event.target.value);
        if (selectedOption) {
            newValue.display = selectedOption.display;
        }
        setValue(newValue);
        dispatch(set({
            id: field.id,
            value: JSON.stringify(newValue),
        }));
    }

    let options = [];
    if (field.id in handlerResults.values) {
        options = handlerResults.values[field.id];
    }

    return (
        <FormControl fullWidth>
            <Typography>Latitude</Typography>
            <InputSlider
                min={-90}
                max={90}
                step={0.1}
                onChange={onChangeLatitude}
                value={value['lat']}
            >
            </InputSlider>
            <Typography>Longitude</Typography>
            <InputSlider
                min={-180}
                max={180}
                step={0.1}
                onChange={onChangeLongitude}
                value={value['lng']}
            >
            </InputSlider>
            <Typography>Locality</Typography>
            <TextField
                fullWidth
                variant="outlined"
                onChange={onChangeLocality}
                style={{ marginBottom: '0.5rem' }}
                value={value['locality']}
            />
            <Typography>Timezone</Typography>
            <Select
                onChange={onChangeTimezone}
                style={{ marginBottom: '0.5rem' }}
                value={value['timezone']}
            >
                {Intl.supportedValuesOf('timeZone').map((zone) => {
                    return <MenuItem value={zone}>{zone}</MenuItem>
                })}
            </Select>
            <Typography>Options for chosen location</Typography>
            <Select
                onChange={onChangeOption}
                value={value['value']}
            >
                {options.map((option) => {
                    return <MenuItem key={option.value} value={option.value}>{option.display}</MenuItem>
                })}
            </Select>
        </FormControl>
    );
}