import { useSelector, useDispatch } from 'react-redux';

import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import useMediaQuery from '@mui/material/useMediaQuery';
import { resetConfig, setConfig } from '../config/actions';
import { set } from '../config/configSlice';
import { setScale } from '../preview/previewSlice';

export default function Controls() {
    const preview = useSelector(state => state.preview);
    const config = useSelector(state => state.config);
    const schema = useSelector(state => state.schema);
    const dispatch = useDispatch();
    const fullWidth = useMediaQuery((theme) => theme.breakpoints.down('sm'));

    let imageType = 'webp';
    if (preview.value.img_type === "gif") {
        imageType = 'gif';
    }

    function downloadPreview() {
        const date = new Date().getTime();
        const element = document.createElement("a");

        // convert base64 to raw binary data held in a string
        let byteCharacters = atob(preview.value.img);

        // create an ArrayBuffer with a size in bytes
        let arrayBuffer = new ArrayBuffer(byteCharacters.length);

        // create a new Uint8Array view
        let uint8Array = new Uint8Array(arrayBuffer);

        // assign the values
        for (let i = 0; i < byteCharacters.length; i++) {
            uint8Array[i] = byteCharacters.charCodeAt(i);
        }

        const file = new Blob([uint8Array], { type: 'image/' + imageType });
        element.href = URL.createObjectURL(file);
        element.download = `tidbyt-preview-${date}.${imageType}`;
        document.body.appendChild(element); // Required for this to work in FireFox
        element.click();
    }

    function downloadConfig() {
        const date = new Date().getTime();
        const element = document.createElement("a");
        const jsonData = config;

        // Use Blob object for JSON
        const file = new Blob([JSON.stringify(jsonData)], { type: 'application/json' });
        element.href = URL.createObjectURL(file);
        element.download = `config-${date}.json`;
        document.body.appendChild(element); // Required for this to work in FireFox
        element.click();
    }

    function selectConfig() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = 'application/json';

        input.onchange = function (event) {
            const file = event.target.files[0];
            if (file.type !== "application/json") {
                return;
            }

            const reader = new FileReader();

            reader.onload = function () {
                let contents = reader.result;
                let json = JSON.parse(contents);
                setConfig(json);
            };

            reader.onerror = function () {
                console.log(reader.error);
            };

            reader.readAsText(file);
        };

        input.click();
    }


    function resetSchema() {
        history.replaceState(null, '', location.pathname);
        resetConfig();
        schema.value.schema.forEach((field) => {
            if (field.default) {
                dispatch(set({
                    id: field.id,
                    value: field.default,
                }));
            };
        });
    };

    const handleScaleChange = (event) => {
        dispatch(setScale(event.target.value === '2'));
    };

    const scaleValue = preview.value.is2x === true ? '2' : '1';

    return (
        <Stack
            sx={{ marginTop: '32px' }}
            spacing={2}
            direction={{ xs: 'column', sm: 'row' }}
            alignItems="flex-start"
            flexWrap="wrap"
        >
            <FormControl size="small" sx={{ minWidth: 160 }} fullWidth={fullWidth}>
                <InputLabel id="render-scale-label">Render Scale</InputLabel>
                <Select
                    labelId="render-scale-label"
                    id="render-scale"
                    value={scaleValue}
                    label="Render Scale"
                    onChange={handleScaleChange}
                >
                    <MenuItem value="1">1x</MenuItem>
                    <MenuItem value="2">2x</MenuItem>
                </Select>
            </FormControl>
            <Button fullWidth={fullWidth} variant="outlined" onClick={() => selectConfig()}>Import Config</Button>
            <Button fullWidth={fullWidth} variant="outlined" onClick={() => downloadConfig()}>Export Config</Button>
            <Button fullWidth={fullWidth} variant="outlined" onClick={() => resetSchema()}>Reset</Button>
            <Button fullWidth={fullWidth} variant="contained" onClick={() => downloadPreview()}>Export Image</Button>
        </Stack>
    );
}
