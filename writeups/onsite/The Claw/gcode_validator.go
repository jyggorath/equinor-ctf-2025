package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

var (
	maxFeedRate      = 5000.0
	maxClawStrength  = 10
	allowedGCommands = map[string]bool{
		"G0": true, // Rapid positioning (X, Y, Z only)
		"G4": true, // Dwell/Wait
	}
	allowedMCommands = map[string]bool{
		"M106": true, // Claw control (S0=open, S1-S10=close)
	}
)

type GCodeValidationError struct {
	Line    int
	Code    string
	Message string
}

func (e GCodeValidationError) Error() string {
	return fmt.Sprintf("Line %d (%s): %s", e.Line, e.Code, e.Message)
}

func (app *App) validateGCode(gcodeText string) ([]string, error) {
	lines := strings.Split(strings.TrimSpace(gcodeText), "\n")
	validatedLines := make([]string, 0)

	// ALWAYS start with G28 XYZ to reset the claw position
	validatedLines = append(validatedLines, "G28 XYZ")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
			continue
		}

		// Validate and correct line
		correctedLine, err := app.validateAndCorrectGCodeLine(line, i+1)
		if err != nil {
			log.Printf("G-code validation failed on line %d: %q - Error: %v", i+1, line, err)
			return nil, err
		}

		if correctedLine != line {
			log.Printf("G-code corrected on line %d: %q -> %q", i+1, line, correctedLine)
		}

		validatedLines = append(validatedLines, correctedLine)
	}

	if len(validatedLines) <= 1 { // Only G28 XYZ
		return nil, fmt.Errorf("no valid G-code commands found (besides mandatory G28 XYZ reset)")
	}

	if len(validatedLines) > 101 { // +1 for mandatory G28 XYZ
		return nil, fmt.Errorf("too many G-code commands (max 100 user commands plus mandatory G28 XYZ, got %d total)", len(validatedLines))
	}

	return validatedLines, nil
}

func (app *App) validateAndCorrectGCodeLine(line string, lineNum int) (string, error) {
	originalLine := line

	// Remove inline comments
	if idx := strings.Index(line, ";"); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}
	if idx := strings.Index(line, "//"); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}

	if line == "" {
		return originalLine, nil
	}

	// Parse command
	parts := strings.Fields(strings.ToUpper(line))
	if len(parts) == 0 {
		return originalLine, nil
	}

	command := parts[0]

	// Validate command type and get corrected parameters
	if strings.HasPrefix(command, "G") {
		if !allowedGCommands[command] {
			return "", GCodeValidationError{
				Line:    lineNum,
				Code:    command,
				Message: fmt.Sprintf("G-code command %s is not allowed", command),
			}
		}
		correctedParams, err := app.validateAndCorrectGCommand(command, parts[1:], lineNum)
		if err != nil {
			return "", err
		}
		// Rebuild the line with corrected parameters
		return command + " " + strings.Join(correctedParams, " "), nil
	} else if strings.HasPrefix(command, "M") {
		if !allowedMCommands[command] {
			return "", GCodeValidationError{
				Line:    lineNum,
				Code:    command,
				Message: fmt.Sprintf("M-code command %s is not allowed", command),
			}
		}
		correctedParams, err := app.validateAndCorrectMCommand(command, parts[1:], lineNum)
		if err != nil {
			return "", err
		}
		// Rebuild the line with corrected parameters
		return command + " " + strings.Join(correctedParams, " "), nil
	} else {
		return "", GCodeValidationError{
			Line:    lineNum,
			Code:    command,
			Message: "Unknown command type (must start with G or M)",
		}
	}
}

func (app *App) validateAndCorrectGCommand(command string, params []string, lineNum int) ([]string, error) {
	correctedParams := make([]string, 0, len(params))

	switch command {
	case "G28": // Home command
		// G28 can have optional axis parameters (X, Y, Z)
		for _, param := range params {
			if len(param) < 1 {
				continue
			}
			paramType := param[0]
			// For G28, we only check if it's a valid axis letter
			if paramType != 'X' && paramType != 'Y' && paramType != 'Z' {
				return nil, GCodeValidationError{
					Line:    lineNum,
					Code:    command,
					Message: fmt.Sprintf("Invalid G28 parameter: %s (only X, Y, Z axes allowed)", param),
				}
			}
			// Parameter is valid, add it as-is
			correctedParams = append(correctedParams, param)
		}
		return correctedParams, nil

	case "G0": // Rapid positioning - ONLY X, Y, Z axes allowed, absolute positioning only
		for _, param := range params {
			if len(param) < 2 {
				continue
			}

			paramType := param[0]
			paramValue := param[1:]

			switch paramType {
			case 'X', 'Y', 'Z': // Only these axes are allowed
				value, err := strconv.ParseFloat(paramValue, 64)
				if err != nil {
					return nil, GCodeValidationError{
						Line:    lineNum,
						Code:    command,
						Message: fmt.Sprintf("Invalid %c axis value: %s", paramType, paramValue),
					}
				}

				// Z axis cannot be negative
				if paramType == 'Z' && value < 0 {
					return nil, GCodeValidationError{
						Line:    lineNum,
						Code:    command,
						Message: fmt.Sprintf("Z axis value cannot be negative: %s (Z must be >= 0)", paramValue),
					}
				}

				// Position values are validated - add as-is
				correctedParams = append(correctedParams, param)

			case 'F': // Feed rate - CLAMP TO MAX 5000
				feedRate, err := strconv.ParseFloat(paramValue, 64)
				if err != nil {
					return nil, GCodeValidationError{
						Line:    lineNum,
						Code:    command,
						Message: fmt.Sprintf("Invalid feed rate value: %s", paramValue),
					}
				}
				if feedRate > maxFeedRate {
					// CLAMP to maximum value instead of rejecting
					correctedParam := fmt.Sprintf("F%.0f", maxFeedRate)
					correctedParams = append(correctedParams, correctedParam)
					log.Printf("Clamped feed rate from F%.0f to F%.0f on line %d", feedRate, maxFeedRate, lineNum)
				} else {
					// Feed rate is within limits, add as-is
					correctedParams = append(correctedParams, param)
				}

			default:
				return nil, GCodeValidationError{
					Line:    lineNum,
					Code:    command,
					Message: fmt.Sprintf("Invalid parameter for G0: %s (only X, Y, Z, F allowed)", param),
				}
			}
		}
		return correctedParams, nil

	case "G4": // Dwell/Wait command
		for _, param := range params {
			if len(param) < 2 {
				continue
			}

			paramType := param[0]
			paramValue := param[1:]

			if paramType == 'P' {
				dwellTime, err := strconv.ParseFloat(paramValue, 64)
				if err != nil {
					return nil, GCodeValidationError{
						Line:    lineNum,
						Code:    command,
						Message: fmt.Sprintf("Invalid dwell time: %s", paramValue),
					}
				}
				if dwellTime > 10000 { // Max 10 seconds - still reject excessive dwell times
					return nil, GCodeValidationError{
						Line:    lineNum,
						Code:    command,
						Message: fmt.Sprintf("Dwell time P%.0f exceeds maximum 10000ms", dwellTime),
					}
				}
				// Dwell time is valid, add as-is
				correctedParams = append(correctedParams, param)
			} else {
				return nil, GCodeValidationError{
					Line:    lineNum,
					Code:    command,
					Message: fmt.Sprintf("Invalid parameter for G4: %s (only P allowed for dwell time)", param),
				}
			}
		}
		return correctedParams, nil

	default:
		return nil, GCodeValidationError{
			Line:    lineNum,
			Code:    command,
			Message: fmt.Sprintf("G-code command %s is not allowed", command),
		}
	}
}

func (app *App) validateAndCorrectMCommand(command string, params []string, lineNum int) ([]string, error) {
	correctedParams := make([]string, 0, len(params))

	switch command {
	case "M106": // Claw control - SECURE VALIDATION WITH CLAMPING
		// Must have exactly one S parameter
		if len(params) != 1 {
			return nil, GCodeValidationError{
				Line:    lineNum,
				Code:    command,
				Message: "M106 must have exactly one S parameter (S0=open, S1-S10=close)",
			}
		}

		param := params[0]
		if len(param) < 2 || param[0] != 'S' {
			return nil, GCodeValidationError{
				Line:    lineNum,
				Code:    command,
				Message: "M106 must have S parameter (S0=open, S1-S10=close)",
			}
		}

		strengthStr := param[1:]
		strength, err := strconv.Atoi(strengthStr)
		if err != nil {
			return nil, GCodeValidationError{
				Line:    lineNum,
				Code:    command,
				Message: fmt.Sprintf("Invalid claw strength value: %s", strengthStr),
			}
		}

		if strength > maxClawStrength {
			correctedParams = append(correctedParams, fmt.Sprintf("S%d", uint8(maxClawStrength)))
			log.Printf("Clamped claw strength from S%d to S%d on line %d", strength, maxClawStrength, lineNum)
		} else {
			correctedParams = append(correctedParams, fmt.Sprintf("S%d", uint8(strength)))
		}

		return correctedParams, nil

	default:
		return nil, GCodeValidationError{
			Line:    lineNum,
			Code:    command,
			Message: fmt.Sprintf("M-code command %s is not allowed", command),
		}
	}
}
