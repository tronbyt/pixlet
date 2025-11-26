import '@fontsource/material-icons';

import { createTheme } from '@mui/material/styles';

import { tronbyt } from './colors';

export const theme = createTheme({
    palette: {
        mode: 'dark',
        primary: {
            main: tronbyt.cyan,
        },
        secondary: {
            main: tronbyt.yellow,
        },
        text: {
            primary: tronbyt.base1,
            secondary: tronbyt.base0,
        },
        background: {
            paper: tronbyt.base03,
            default: tronbyt.base02,
        },
    },
    typography: {
        fontFamily: [
            '-apple-system',
            'BlinkMacSystemFont',
            '"Segoe UI"',
            'Roboto',
            '"Helvetica Neue"',
            'Arial',
            'sans-serif',
            '"Apple Color Emoji"',
            '"Segoe UI Emoji"',
            '"Segoe UI Symbol"',
        ].join(','),
    },
});
