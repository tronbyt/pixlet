import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';

import Logo from './logo.svg?react';
import { tronbyt } from '../theme/colors';


export default function NavBar() {
    return (
        <AppBar sx={{ backgroundColor: tronbyt.base02, backgroundImage: 'none' }} position="static" enableColorOnDark>
            <Toolbar disableGutter>
                <Logo style={{ maxHeight: '32px' }} />
                <Typography variant="h6" sx={{ mx: 2, fontWeight: 700 }}>
                    Pixlet
                </Typography>
            </Toolbar>
        </AppBar>
    )
}
