import { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { set } from './configSlice';
import { loading } from './paramSlice';
import { setScale, setTimezone, setLocale } from '../preview/previewSlice';


export default function ParamSetter() {
    const params = new URLSearchParams(document.location.search);
    const dispatch = useDispatch();

    useEffect(() => {
        const renderScale = params.get('_renderScale');
        if (renderScale === '1' || renderScale === '2') {
            dispatch(setScale(renderScale === '2'));
        }

        params.forEach((value, key) => {
            if (key === '_renderScale') {
                return;
            }
            if (key === '_metaTimezone') {
                dispatch(setTimezone(value));
                return;
            }
            if (key === '_metaLocale') {
                dispatch(setLocale(value));
                return;
            }
            dispatch(set({
                id: key,
                value: value,
            }));
        });
        dispatch(loading(false));
    }, []);

    return null;
};
