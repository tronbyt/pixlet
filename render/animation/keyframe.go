package animation

// Keyframe defines a specific point in time in the animation.
//
// The keyframe _percentage_ can is expressed as a floating point value between `0.0` and `1.0`.
type Keyframe struct {
	// Percentage of the time at which this keyframe occurs through the animation.
	Percentage Percentage `starlark:"percentage,required"`
	// List of transforms at this keyframe to interpolate to or from.
	Transforms []Transform `starlark:"transforms,required"`
	// Easing curve to use, default is 'linear'.
	Curve Curve `starlark:"curve"`
}
