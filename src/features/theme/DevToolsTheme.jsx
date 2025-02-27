import ThemeProvider from '@mui/system/ThemeProvider';

import { theme } from './theme';
import './styles.css';


export default function DevToolsTheme(props) {
    return (
        <ThemeProvider theme={theme}>
            {props.children}
        </ThemeProvider>
    );
}