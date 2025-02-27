// Largely based on https://mui.com/material-ui/react-slider/#InputSlider.js
import React from 'react';

import { styled } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid2';
import Slider from '@mui/material/Slider';
import MuiInput from '@mui/material/Input';

const Input = styled(MuiInput)`
  width: 80px;
`;

export default function InputSlider({ min, max, step, value, onChange}) {
  const handleSliderChange = (event, newValue) => {
    onChange({ ...event, target: { ...event.target, value: newValue } });
  };

  const handleInputChange = (event) => {
    let newValue = event.target.value === '' ? '' : Number(event.target.value);
    if (newValue < min) {
      newValue = min;
    } else if (newValue > max) {
      newValue = max;
    }
    onChange({ ...event, target: { ...event.target, value: newValue } });
  };

  const handleBlur = () => {
    if (value < min) {
      setValue(min);
    } else if (value > max) {
      setValue(max);
    }
  };

  return (
    <Box sx={{ width: 250 }}>
      <Grid container spacing={2} alignItems="center">
        <Grid size="grow">
          <Slider
            value={value}
            min={min}
            max={max}
            step={step}
            onChange={handleSliderChange}
            aria-labelledby="input-slider"
          />
        </Grid>
        <Grid>
          <Input
            value={value}
            size="small"
            onChange={handleInputChange}
            onBlur={handleBlur}
            inputProps={{
              step: {step},
              min: {min},
              max: {max},
              type: 'number',
              'aria-labelledby': 'input-slider',
            }}
          />
        </Grid>
      </Grid>
    </Box>
  );
}