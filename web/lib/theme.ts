import {
  BrandVariants,
  createLightTheme,
  createDarkTheme,
} from '@fluentui/react-components';

// Generate 16-stop palette from #EB5424
const gobackupBrand: BrandVariants = {
  10: '#1a0500',
  20: '#3d0e00',
  30: '#5c1600',
  40: '#7a1e00',
  50: '#992600',
  60: '#b82e00',
  70: '#d63600',
  80: '#EB5424', // Primary brand color
  90: '#ef6b42',
  100: '#f28260',
  110: '#f5997e',
  120: '#f8b09c',
  130: '#fac7ba',
  140: '#fdded8',
  150: '#fef5f2',
  160: '#fffaf8',
};

export const gobackupLightTheme = {
  ...createLightTheme(gobackupBrand),
};

export const gobackupDarkTheme = {
  ...createDarkTheme(gobackupBrand),
};
