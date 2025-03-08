import { useState, useEffect, useCallback } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import Typography from '@mui/material/Typography';
import FormControl from '@mui/material/FormControl';

import { set } from '../../../config/configSlice';

// Command components for the location picker - these create the UI for the dropdown interface
// Container component that wraps the entire command interface
const Command = ({ children, className }) => (
    <div className={`relative rounded-lg ${className}`}>
        {children}
    </div>
);

// Input field component for search queries
const CommandInput = ({ placeholder, value, onValueChange }) => (
    <input
        className="w-full px-4 py-2 rounded-lg border focus:outline-none focus:ring-2"
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
    />
);

// Container for the dropdown list of location results
const CommandList = ({ children }) => (
    <div className="mt-1 max-h-60 overflow-auto rounded-lg border bg-white">
        {children}
    </div>
);

// Individual item in the dropdown list that user can select
const CommandItem = ({ children, onSelect }) => (
    <div
        className="px-4 py-2 hover:bg-gray-100 cursor-pointer flex items-center"
        onClick={onSelect}
    >
        {children}
    </div>
);

// Helper function to extract URL parameters
const getUrlParams = () => {
    // Parse the query string from the URL
    const queryString = window.location.search;
    const params = new URLSearchParams(queryString);

    // Create an object with all parameters
    const urlParams = {};
    for (const [key, value] of params.entries()) {
        urlParams[key] = value;
    }

    return urlParams;
};

// Main location picker component that integrates with Geoapify API
const LocationPicker = ({ onLocationSelect, initialLocation }) => {
    // State for search input, results, loading status, and dropdown visibility
    const [query, setQuery] = useState(initialLocation?.locality || '');
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [isOpen, setIsOpen] = useState(false);

    // Debounce function to prevent excessive API calls during typing
    const debounce = (func, wait) => {
        let timeout;
        return (...args) => {
            clearTimeout(timeout);
            timeout = setTimeout(() => func(...args), wait);
        };
    };

    // Function to fetch location results from Geoapify API
    const searchLocations = async (searchQuery) => {
        if (!searchQuery) {
            setResults([]);
            setIsOpen(false); // Close dropdown if query is empty
            return;
        }

        setLoading(true);
        setError(null);

        try {
            const apiKey = '49863bd631b84a169b347fafbf128ce6';
            const encodedQuery = encodeURIComponent(searchQuery);
            const response = await fetch(
                `https://api.geoapify.com/v1/geocode/search?text=${encodedQuery}&apiKey=${apiKey}`
            );

            if (!response.ok) {
                throw new Error('Failed to fetch locations');
            }

            const data = await response.json();

            // Transform API response to match our format
            const locations = data.features.map(feature => ({
                formattedName: feature.properties.formatted,
                locality: feature.properties.formatted,
                lat: feature.properties.lat,
                lng: feature.properties.lon,
                timezone: feature.properties.timezone?.name || 'America/New_York'
            }));

            setResults(locations);
        } catch (err) {
            setError('Failed to load locations');
            console.error('Error fetching locations:', err);
        } finally {
            setLoading(false);
        }
    };

    // Create a debounced version of the search function (300ms delay)
    const debouncedSearch = useCallback(
        debounce(searchLocations, 300),
        []
    );

    // Handler for search input changes
    const handleSearch = (value) => {
        setQuery(value);
        setIsOpen(true); // Open dropdown when searching
        debouncedSearch(value);
    };

    // Close dropdown when clicking outside the component
    useEffect(() => {
        const handleClickOutside = () => {
            setIsOpen(false);
        };

        // Add event listener when component mounts
        document.addEventListener('click', handleClickOutside);

        // Return cleanup function
        return () => {
            document.removeEventListener('click', handleClickOutside);
        };
    }, []);

    // Prevent dropdown from closing when clicking inside the component
    const handleCommandClick = (e) => {
        e.stopPropagation();
    };

    return (
        <Command className="w-full" onClick={handleCommandClick}>
            <CommandInput
                placeholder="Search for a location..."
                value={query}
                onValueChange={handleSearch}
                onClick={() => setIsOpen(true)} // Open dropdown when clicking input
            />
            {isOpen && loading && (
                <div className="p-4 text-center text-gray-500">
                    Loading...
                </div>
            )}
            {isOpen && error && (
                <div className="p-4 text-center text-red-500">
                    {error}
                </div>
            )}
            {isOpen && results.length > 0 && (
                <CommandList>
                    {results.map((location, index) => (
                        <CommandItem
                            key={index}
                            onSelect={() => {
                                onLocationSelect(location);
                                setIsOpen(false); // Close dropdown after selection
                                setResults([]); // Clear results
                            }}
                        >
                            <span className="mr-2">📍</span>
                            {location.formattedName}
                        </CommandItem>
                    ))}
                </CommandList>
            )}
        </Command>
    );
};

// Main exported component that wraps the LocationPicker with Redux integration
export default function LocationForm({ field }) {
    // Get URL parameters to check for location data
    const urlParams = getUrlParams();

    // Check if location parameter exists in URL (it's a JSON string)
    const urlLocationParam = urlParams.location;
    let urlLocation = null;

    // Try to parse the URL location parameter if it exists
    if (urlLocationParam) {
        try {
            urlLocation = JSON.parse(decodeURIComponent(urlLocationParam));
        } catch (err) {
            console.error('Error parsing location from URL:', err);
        }
    }

    // Determine default values based on defaults, field props, and URL params
    const getDefaultValues = () => {
        const defaultValues = {
            // Default to Brooklyn as fallback
            'lat': 40.678,
            'lng': -73.944,
            'locality': 'Brooklyn, New York',
            'timezone': 'America/New_York',
        };

        // Override with field.default if available
        if (field.default) {
            Object.assign(defaultValues, field.default);
        }

        // Override with URL location if available
        if (urlLocation) {
            return urlLocation; // Replace the entire location object
        }

        return defaultValues;
    };

    // State for location value and component initialization tracking
    const [value, setValue] = useState(getDefaultValues());
    const [initialized, setInitialized] = useState(false);

    // Redux hooks for state management
    const config = useSelector(state => state.config);
    const dispatch = useDispatch();

    // Set component as initialized on first render
    useEffect(() => {
        // Set initialized to true immediately - we don't need to search
        // since we're getting the full location object from the URL
        setInitialized(true);
    }, []);

    // Load location from Redux store or initialize with defaults
    useEffect(() => {
        if (initialized && field.id in config) {
            // If location exists in Redux, use that value
            setValue(JSON.parse(config[field.id].value));
        } else if (initialized) {
            // Otherwise use default values and save to Redux
            const defaultValues = getDefaultValues();
            dispatch(set({
                id: field.id,
                value: JSON.stringify(defaultValues),
            }));
            setValue(defaultValues);
        }
    }, [config, dispatch, field.id, field.default, initialized]);

    // Handler for when user selects a location from the dropdown
    const handleLocationSelect = (location) => {
        const newValue = {
            lat: location.lat,
            lng: location.lng,
            locality: location.locality || location.formattedName,
            timezone: location.timezone
        };

        // Update local state and Redux store
        setValue(newValue);
        dispatch(set({
            id: field.id,
            value: JSON.stringify(newValue),
        }));
    };

    return (
        <FormControl fullWidth>
            <Typography variant="subtitle1" gutterBottom>Location</Typography>

            {/* Location search component */}
            <LocationPicker
                onLocationSelect={handleLocationSelect}
                initialLocation={value}
            />

            {/* Display the currently selected location details */}
            {value && (
                <div className="mt-4 p-3 border rounded bg-gray-50">
                    <Typography variant="body2"><strong>Selected:</strong> {value.locality}</Typography>
                    <Typography variant="body2"><strong>Coordinates:</strong> {value.lat.toFixed(3)}, {value.lng.toFixed(3)}</Typography>
                    <Typography variant="body2"><strong>Timezone:</strong> {value.timezone}</Typography>
                </div>
            )}
        </FormControl>
    );
}