import { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { Sketch } from '@uiw/react-color';
import { set } from '../../config/configSlice';


export default function Color({ field }) {
    const [color, setColor] = useState(field.default || '#000');
    // TODO: expose the color palette specified in the schema.
    const config = useSelector(state => state.config);
    const dispatch = useDispatch();

    useEffect(() => {
        if (field.id in config) {
            setColor(config[field.id].value);
        } else if (field.default) {
            dispatch(set({
                id: field.id,
                value: field.default,
            }));
        }
    }, [config]);

    const onChange = (color) => {
        setColor(color.hex);

        // Skip updates that contain an error.
        if (color.hasOwnProperty("error")) {
            return;
        }

        dispatch(set({
            id: field.id,
            value: color.hex,
        }));
    };

    return (
        <Sketch color={color} onChange={onChange} disableAlpha />
    );
}
