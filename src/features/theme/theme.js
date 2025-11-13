import '@fontsource/material-icons';

import { createTheme } from '@mui/material/styles';

import { solarized } from './colors';

export const theme = createTheme({
    palette: {
        mode: 'light',
        primary: {
            main: solarized.cyan,
        },
        secondary: {
            main: solarized.yellow,
        },
        text: {
            primary: solarized.base1,
            secondary: solarized.base0,
        },
        background: {
            paper: solarized.base02,
            default: solarized.base02,
        },
    },
    components: {
        MuiSvgIcon: {
            defaultProps: {
                htmlColor: solarized.base1,
                color: solarized.base1,
            },
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
