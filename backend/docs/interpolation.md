# Data Interpolation Process

This document describes the process of interpolating missing data points within datasets that represent measurements over time at fixed geographic locations. The goal is to fill in gaps (`NaN`) based on surrounding valid data points according to specific rules using different interpolation strategies.

## Description

The application provides two main strategies for interpolating missing data:

1.  **Linear Interpolation (`interpolateLinearyDataRow`)**: This function processes a sequence of data points for a _single_ geographic location, ordered by their measurement time. It identifies missing values within this temporal sequence and fills them based on neighboring data points within the same location's timeline.

    `interpolateLinearyDataRow()` applies the following rules:

    - **Direct Neighbors Interpolation:** If a data point at time `T` is missing, and the points at `T-1` and `T+1` for the _same location_ are valid, the missing value at `T` is filled with the average of the values at `T-1` and `T+1`.
    - **Linear Interpolation for Longer Gaps:** If a data point at time `T` is missing and `T-1` is valid, the function searches for the next valid data point at time `T+n` for the same location. If found, all missing values between `T` and `T+n-1` are filled using linear interpolation between the valid values at `T-1` and `T+n`.
    - **Unbounded Gaps Remain Unfilled:** Gaps at the beginning or end of a location's record, or gaps not bounded by valid data points on both sides, will not be interpolated and remain as missing (`NaN`).

2.  **Area-Based Interpolation (`interpolateDataArea`)**: This function processes a 2D grid of data points representing measurements for a _single_ timestamp across multiple geographic locations. It identifies contiguous groups of missing values (`NaN`) and fills them based on the average of surrounding valid data points within that grid.

    `interpolateDataArea()` applies the following rules:

    - **Surrounded NaN Groups Interpolation:** A contiguous group of `NaN` values is filled _only if_ the entire group is completely surrounded by non-`NaN` values within the 2D grid. This means no `NaN` within the group is on the edge of the data grid or adjacent to another `NaN` that is itself on the edge.
    - **Average of Surrounding Data:** For a surrounded group of `NaN`s, all `NaN`s within the group are filled with the average of _all_ the non-`NaN` values directly adjacent (including diagonals) to _any_ `NaN` within that group.
    - **Unsurrounded NaNs Remain Unfilled:** Any `NaN` value or group of `NaN` values that is not completely surrounded by valid data points (e.g., they are on the edge of the grid) will _not_ be interpolated and will remain as missing (`NaN`).

Both `interpolateLinearyDataRow` and `interpolateDataArea` functions work with data structures that implement the `InterpolatableData` interface, allowing them to be applied to various datasets within the project.

## `InterpolatableData` Interface

The `InterpolatableData` interface defines the contract for any data point that can be processed by the interpolation logic. It requires two methods:

```go
type InterpolatableData interface {
	Value() float32     // Returns the data value (e.g., chlorophyll concentration)
	SetValue(float32)   // Sets the data value
}
```

This interface allows the interpolation functions to operate on the values without knowing the specific details of the underlying struct (like `models.ChlorophyllData`).

## How to Add New Data to be Interpolated

This section outlines the steps to enable interpolation for a new data source, assuming the data source has already been integrated into the project following the instructions in `adding_new_data.md`.

1.  **Implement Database Queries:**

    - In a file (`database/queries-source_name.go`), write SQL queries and corresponding Go functions within the `database.Service` implementation to:
      - Find relevant data for interpolation (e.g., unique geographic locations for linear interpolation, or timestamps for area-based interpolation).
      - Retrieve the necessary data points (either a time series for a location or a 2D grid for a timestamp).
      - Update the data value for specific records based on their unique identifiers (ID).

2.  **Implement `InterpolatableData` Interface:**

    - Ensure the Go struct that holds the data for your new source implements the `InterpolatableData` interface. This involves adding the `Value()` and `SetValue(float32)` methods to the struct.

3.  **Implement Interpolation Method(s) in `Interpolator`:**

    - In `internal/interpolator/interpolator.go`, add one or more new methods (e.g., `RunSourceNameLinearInterpolation`, `RunSourceNameAreaInterpolation`) to the `Interpolator` struct. These methods will orchestrate the interpolation process for your specific data source using the appropriate interpolation function(s):
      - Call the database function(s) to retrieve the data in the required format (slice for linear, 2D slice for area).
      - Convert the retrieved data points into the appropriate slice(s) of `InterpolatableData`.
      - Call the relevant interpolation function (`interpolateLinearyDataRow()` or `interpolateDataArea()`) with the prepared data.
      - Iterate through the modified data and update the database for any values that were filled.

4.  **Integrate into the Updater:**
    - In `internal/scheduler/updater.go` find the `update` method that handles the daily data process for your new source.
    - After the data for this source has been successfully downloaded and saved to the database, call the relevant interpolation method(s) you created in the previous step to perform the gap filling.

## How to Create New Interpolation Functions

The system is designed to be extensible with new interpolation algorithms.

1.  Define a new function that takes a slice of `InterpolatableData` (or a similar interface if your new method requires different capabilities, like the 2D slice for `interpolateDataArea`) and applies your desired interpolation algorithm.
2.  Integrate this new function into the `Interpolator` (or a new interpolation service) and expose it through appropriate methods.
3.  Update the relevant data processing workflows to use your new interpolation function.
4.  Preferably implement tests to guarantee correct functioning of new method.
