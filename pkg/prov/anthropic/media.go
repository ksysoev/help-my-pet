package anthropic

// mediaExtractionPrompt - enhanced to cover more scenarios
const mediaExtractionPrompt = `Extract and describe all objective information from the provided media. Determine the media type first, then focus ONLY on visible elements and factual information without making diagnoses.

IF PET PHOTO:
Observe and document the following elements:
1. Subject identification
   - Species and breed characteristics (be specific about visible traits rather than guessing exact breeds)
   - Approximate age indicators (teeth visibility, gray hair, muscle tone, eye clarity)
   - Size estimation relative to visible reference objects (provide metric estimates when possible)
   - Color patterns and distinctive markings (be precise with color terminology and distribution patterns)

2. Physical condition assessment
   - Body condition (thin, ideal, overweight - note visible indicators like rib visibility, waist definition)
   - Posture and positioning (natural vs. unnatural, weight-bearing on all limbs, symmetry)
   - Coat/skin condition (texture, shininess, hair loss areas, dandruff, parasites if visible)
   - Visible anatomical features (proportions, visible muscle tone, joint appearance)

3. Clinical observations
   - Any visible abnormalities (exact location, precise size with estimates, shape, color, texture)
   - Presence of discharge, swelling, or lesions (description of color, consistency, and extent)
   - Symmetry or asymmetry of features (specify which features and how they differ)
   - Signs of discomfort (facial expression, body position, tension indicators)

4. Environmental context
   - Setting (indoor/outdoor, specific features of the environment)
   - Relevant objects in proximity to the animal (potential hazards, items of interest)
   - Lighting conditions (bright, dim, natural, artificial - how it might affect observation)
   - Surface the animal is on (material, cleanliness, stability)

5. Image quality factors
   - Clarity and focus (specify areas that are clear vs. blurry)
   - Lighting adequacy (shadows, overexposure, color distortion)
   - Angle and perspective (which body parts are visible vs. hidden)
   - Presence of measurement references (rulers, common objects for scale)
   - Capture quality limitations (motion blur, obstructions, partial views)

6. Behavioral indicators (if animal is active in image)
   - Body language (ear position, tail position, stance, facial expressions)
   - Interaction with environment or other animals/people
   - Activity level (resting, playing, alert, stressed)
   - Signs of specific behaviors (territorial, playful, fearful, aggressive)

IF LAB TEST RESULTS:
Extract the following elements:
1. Report identification
   - Test type(s) (complete blood count, chemistry panel, urinalysis, etc.)
   - Date of testing/reporting (collection date vs. report date if both available)
   - Patient identifiers (name, ID number, species, breed, age, sex)
   - Requesting veterinarian or clinic information

2. Test parameters and results
   - Parameter name (full name, not just abbreviations)
   - Measured value (with exact decimal precision as shown)
   - Units of measurement (standard or alternative units)
   - Reference/normal range (with any age/sex/breed specific notations)
   - Flags for abnormal values (H, L, etc. and their meaning if provided)
   - Percentage deviation from reference range midpoint when applicable

3. Sample information
   - Sample type (blood, urine, tissue, etc. and specific type if indicated)
   - Collection date/time (include time if available)
   - Sample quality notes (hemolysis, lipemia, inadequate volume, etc.)
   - Preservation method if mentioned

4. Additional report elements
   - Any comments or notes by lab personnel or pathologists
   - Quality control indicators or verification notices
   - Testing methodology mentioned (machine types, manual differentials)
   - Any disclaimers present about result interpretation
   - Recommendations for follow-up testing if present

IF PET FOOD/NUTRITION IMAGE:
Extract the following elements:
1. Product identification
   - Brand name and specific product line
   - Food type (dry kibble, wet food, treats, supplements)
   - Target species and life stage (puppy, adult, senior)
   - Special dietary purpose if indicated (weight management, sensitive stomach)

2. Nutritional information
   - Macronutrient percentages (protein, fat, fiber, moisture)
   - Caloric content (kcal per cup/can/serving)
   - Ingredient list in order of appearance
   - Guaranteed analysis values
   - Special nutritional claims (grain-free, limited ingredients)

3. Feeding guidelines
   - Recommended portions based on weight/age
   - Feeding frequency recommendations
   - Transition instructions if present
   - Storage instructions

4. Regulatory information
   - AAFCO statement and feeding trial information
   - Country of manufacture
   - Expiration or best-by date
   - Lot number or manufacturing codes

IF TRAINING/BEHAVIORAL DOCUMENTATION:
Extract the following elements:
1. Training setup
   - Training environment (home, yard, training facility, public space)
   - Equipment visible (leashes, clickers, treat pouches, barriers)
   - Distractions present (other animals, people, noises)
   - Safety measures in place

2. Behavioral observation
   - Animal's body language (ear position, tail carriage, stance)
   - Focus direction (on handler, environment, specific item)
   - Signs of stress or comfort (panting, relaxed muscles, tension)
   - Response to cues or commands if demonstrated
   - Interaction with handlers, other animals, or objects

3. Training methodology indicators
   - Reinforcement types visible (treats, toys, praise)
   - Training aids being used (head halters, harnesses, long lines)
   - Training approach apparent from context (luring, shaping, capturing)
   - Stage of training (introduction of concept, practice, proofing)

IF PET CARE EQUIPMENT/SETUP:
Extract the following elements:
1. Housing/containment
   - Type of enclosure (crate, pen, aquarium, terrarium, cage)
   - Dimensions and space adequacy relative to animal size
   - Materials and construction quality
   - Safety features and potential hazards

2. Environmental enrichment
   - Bedding/substrate type and condition
   - Toys and mental stimulation items
   - Climbing/exercise equipment
   - Hiding spots and comfort areas

3. Maintenance and hygiene
   - Cleanliness level of habitat/equipment
   - Water and food container types and placement
   - Waste management systems
   - Ventilation and light sources

4. Pet-specific accommodations
   - Temperature regulation equipment (heaters, fans, thermometers)
   - Specialized equipment for species needs
   - Accessibility features for young, elderly, or disabled pets
   - Safety modifications for the environment

IF VETERINARY/MEDICAL DOCUMENTS:
Extract the following elements:
1. Document type and identification
   - Type of record (vaccination, prescription, discharge instructions)
   - Issuing veterinarian/clinic information
   - Date of issue and expiration dates if applicable
   - Patient identification information

2. Medical information
   - Diagnoses or conditions mentioned
   - Treatments prescribed or performed
   - Medication names, dosages, and administration instructions
   - Follow-up recommendations and schedules

3. Preventative care information
   - Vaccination types and dates administered/due
   - Parasite prevention products and schedules
   - Screening tests recommended or performed
   - Wellness recommendations

Be precise, factual, and avoid interpretations or diagnoses.
Include measurements when possible using visible references.
Document colors, textures, specific locations, and values accurately.
If certain information is not visible or available, explicitly state what could not be observed rather than making assumptions.`

// mediaOutputFormat - expanded to cover additional scenarios
const mediaOutputFormat = `First, identify the type of media provided and then provide a comprehensive description as follows:

For pet photos: Describe the subject's species and apparent breed characteristics, noting specific physical traits that inform this assessment. Estimate the animal's age based on visible indicators such as teeth condition, muscle tone, and coat appearance. Note the animal's approximate size using any visible reference objects for scale. Document all coloration patterns and distinctive markings with precise terminology. Assess body condition as thin, ideal, or overweight and provide the specific visual indicators that support this assessment. Describe the animal's posture, how weight is distributed, and any asymmetry in positioning. Detail the coat and skin condition including texture, shine, and any areas of hair loss or abnormality. Note any visible anatomical features of interest. Document all abnormalities with precise locations, size estimates, shape, color, and texture characteristics. Note any discharge, swelling, or lesions with descriptions of color, consistency, and extent. Identify any asymmetry in physical features. Document indicators of potential discomfort such as facial expressions or body positioning. Describe the setting (indoor/outdoor) and any relevant objects near the animal. Note lighting conditions and the surface the animal is on. Comment on image quality factors including clarity, lighting adequacy, available angles, and any limitations that might affect assessment accuracy. If the animal is active, describe body language indicators, interaction with the environment, activity level, and any specific behaviors demonstrated.

For lab test results: Identify the test type, report date, patient information (name, ID, species/breed), and requesting clinic. Specify the sample type, collection date, and any notes about sample quality. For each parameter tested, provide the parameter name, measured value, units, reference range, and status (normal/high/low). Include any laboratory comments, testing methods mentioned, and quality control indicators. Note any significant deviations from reference ranges, particularly those flagged by the laboratory. Include any pathologist comments or recommendations for follow-up testing if present.

For pet food/nutrition images: Identify the brand name, product line, food type, and intended pet category. Detail the macronutrient percentages, caloric content, and key ingredients listed. Note any special dietary claims or formulation features. Document feeding guidelines including recommended portions based on pet size and age. Include regulatory information such as AAFCO statements, country of manufacture, and expiration dates if visible. Comment on packaging condition, storage requirements, and any visible quality issues with the food itself if shown.

For training/behavioral documentation: Describe the training environment, equipment visible, and safety measures in place. Detail the animal's body language, focus direction, and signs of stress or comfort. Note any visible training methodology indicators such as reinforcement types, training aids being used, and apparent training approach. Assess the animal's response to handlers, other animals, or commands if demonstrated. Comment on the appropriateness of the training setup for the apparent goal if discernible from the image.

For pet care equipment/setup: Identify the type of housing or containment system shown, including dimensions relative to animal size. Describe materials, construction quality, and safety features. Document environmental enrichment items, bedding types, toys, and exercise equipment. Assess cleanliness, maintenance status, and hygiene considerations. Note specific accommodations for the pet's species needs, age, or health requirements. Identify potential safety concerns or improvements if evident.

For veterinary/medical documents: Identify the document type, issuing clinic, and date. Extract patient identification information, diagnoses or conditions mentioned, and treatments prescribed. Document medication names, dosages, and administration instructions. Note preventative care information including vaccinations, parasite prevention, and wellness recommendations. Include follow-up instructions and scheduling details.

Present all information in a flowing narrative format while maintaining comprehensive detail. If certain information cannot be determined from the media provided, explicitly state what elements could not be assessed rather than making assumptions.

EXAMPLES:

EXAMPLE 1 (Pet Photo):
The image shows a domestic short-haired cat with predominantly orange tabby coloration displaying classic mackerel pattern stripes. The cat appears to be an adult, estimated between 3-7 years based on muscular development and coat condition. Size estimation is difficult without reference objects, but the cat appears to be of average build for the breed, likely 4-5kg. The body condition appears ideal with a visible waistline and appropriate muscle tone; ribs are not visible but likely palpable under the coat. The cat is positioned in a relaxed sitting posture with weight evenly distributed and symmetrical body alignment. The coat appears glossy and well-maintained with no visible bald patches, parasites, or skin abnormalities. No discharge, swelling, or lesions are visible on exposed skin areas. The cat's facial expression appears relaxed with partially closed eyes, suggesting comfort. The animal is photographed indoors on what appears to be a beige fabric sofa in natural lighting coming from a nearby window. The image is clear and well-lit, though the angle only provides a clear view of the animal's left side profile; the right side, underside, and rear portions of the body cannot be assessed in this image. The cat's ears are in a neutral forward position and its tail is wrapped around its body in a relaxed manner, suggesting the animal is calm and comfortable in its environment. There are no other animals or people visible in the image with which the cat is interacting.

EXAMPLE 2 (Lab Test Results):
The media shows a Complete Blood Count (CBC) and Chemistry Panel report dated May 12, 2023 for a patient identified as "Max," a 6-year-old male neutered Golden Retriever (ID #GR-2023-456). The test was requested by Valley Pet Clinic (Dr. Anderson). The sample is identified as whole blood collected on May 11, 2023 at 9:15 AM with no quality issues noted. The CBC results show RBC at 6.8 million/μL (reference: 5.5-8.5) which is within normal range, Hemoglobin at 16.2 g/dL (reference: 12.0-18.0) which is normal, and White Blood Cell count at 14.8 thousand/μL (reference: 5.5-13.9) which is flagged as high (H) at 6.5% above the upper reference limit. Platelets measure 325 thousand/μL (reference: 175-500) which is normal. Chemistry panel results show elevated Alkaline Phosphatase at 210 U/L (reference: 20-150) with an (H) flag, representing a 40% elevation above the upper reference limit. All other chemistry values including ALT, AST, BUN, Creatinine, Glucose, Total Protein, and Albumin fall within their respective reference ranges. A note from the pathologist indicates mild leukocytosis and suggests correlation with clinical signs and possible follow-up tests if inflammation is suspected. The report includes a quality control statement confirming test validity and mentions analysis was performed on a ProCyte DX analyzer with manual differential confirmation.

EXAMPLE 3 (Pet Food Image):
The image shows a package of "Natural Balance L.I.D. Limited Ingredient Diets" dry dog food in the Sweet Potato & Fish Formula variety. This product is marketed for adult dogs with food sensitivities. The packaging identifies this as a limited ingredient diet specifically formulated for dogs with food sensitivities or allergies. The guaranteed analysis states 21.0% crude protein, 10.0% crude fat, 4.5% crude fiber, and 10.0% moisture. Caloric content is listed as 350 kcal per cup. The first five ingredients listed are sweet potatoes, salmon, salmon meal, canola oil, and potato protein. The food claims to be grain-free with no artificial colors, flavors, or preservatives. Feeding guidelines recommend 1½ to 2 cups daily for dogs weighing 10-20 pounds, 2 to 3 cups for dogs 20-40 pounds, and 3 to 4½ cups for dogs 40-60 pounds, with instructions to divide into two meals. The package includes an AAFCO statement confirming the food meets nutritional levels established for maintenance of adult dogs. The package indicates the product is manufactured in the USA with a best-by date of September 15, 2024 and lot number NB22453. Storage instructions recommend keeping the bag sealed and stored in a cool, dry place. The packaging appears intact with no visible damage or quality concerns with the kibble visible in the product image window.

Present all information in a flowing narrative format while maintaining comprehensive detail. If certain information cannot be determined from the media provided, explicitly state what elements could not be assessed rather than making assumptions.`
