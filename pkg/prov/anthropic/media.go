package anthropic

// mediaExtractionPrompt defines instructions for the Haiku model to extract detailed
// information from both pet photos and lab test results
const mediaExtractionPrompt = `Extract and describe all objective information from the provided media (pet photo or lab test results).
Determine the media type first, then focus ONLY on visible elements and factual information without making diagnoses.

IF PET PHOTO:
Observe and document the following elements:
1. Subject identification
   - Species and breed characteristics
   - Approximate age indicators
   - Size estimation relative to visible reference objects
   - Color patterns and distinctive markings

2. Physical condition assessment
   - Body condition (thin, ideal, overweight)
   - Posture and positioning
   - Coat/skin condition (texture, shininess, hair loss areas)
   - Visible anatomical features

3. Clinical observations
   - Any visible abnormalities (location, size, shape, color)
   - Presence of discharge, swelling, or lesions
   - Symmetry or asymmetry of features
   - Signs of discomfort (facial expression, body position)

4. Environmental context
   - Setting (indoor/outdoor)
   - Relevant objects in proximity to the animal
   - Lighting conditions
   - Surface the animal is on

5. Image quality factors
   - Clarity and focus
   - Lighting adequacy
   - Angle and perspective
   - Presence of measurement references

IF LAB TEST RESULTS:
Extract the following elements:
1. Report identification
   - Test type(s)
   - Date of testing/reporting
   - Patient identifiers (name, ID number, species)
   - Requesting veterinarian or clinic

2. Test parameters and results
   - Parameter name
   - Measured value
   - Units of measurement
   - Reference/normal range
   - Flags for abnormal values (H, L, etc.)

3. Sample information
   - Sample type (blood, urine, tissue, etc.)
   - Collection date/time
   - Sample quality notes

4. Additional report elements
   - Any comments or notes by lab personnel
   - Quality control indicators
   - Testing methodology mentioned
   - Any disclaimers present

Be precise, factual, and avoid interpretations or diagnoses.
Include measurements when possible using visible references.
Document colors, textures, specific locations, and values accurately.`

// mediaOutputFormat defines the structured output format for both photo and lab test information
const mediaOutputFormat = `First, identify the type of media provided and then use the appropriate format below.

FOR PET PHOTOS:
SUBJECT:
- Species and breed: [description]
- Age estimate: [description]
- Size: [description]
- Coloration and markings: [description]

PHYSICAL CONDITION:
- Body condition: [thin/ideal/overweight]
- Posture and positioning: [description]
- Coat/skin condition: [description]
- Notable features: [description]

CLINICAL OBSERVATIONS:
- Visible abnormalities: [location, description, measurements]
- Symmetry issues: [description]
- Any discharge or lesions: [description]
- Signs of discomfort: [description]

ENVIRONMENT:
- Setting: [indoor/outdoor, description]
- Relevant objects nearby: [description]
- Animal's behavior in environment: [description]

IMAGE QUALITY:
- Clarity and focus: [description]
- Lighting conditions: [description]
- Available angles/perspectives: [description]
- Assessment limitations: [description]

FOR LAB TEST RESULTS:
REPORT IDENTIFICATION:
- Test type: [description]
- Report date: [date]
- Patient information: [animal name, ID, species/breed]
- Requesting clinic: [name]

SAMPLE INFORMATION:
- Sample type: [blood, urine, etc.]
- Collection date: [date]
- Sample condition: [any notes about sample quality]

TEST RESULTS:
[For each parameter tested, include:]
- Parameter: [name]
- Value: [measured value]
- Units: [units]
- Reference range: [min-max]
- Status: [normal/high/low]

ADDITIONAL NOTES:
- Lab comments: [any comments provided]
- Testing methods: [any methodology noted]
- Quality indicators: [any QC information]`
