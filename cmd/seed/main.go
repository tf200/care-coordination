package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"care-cordination/lib/config"
	db "care-cordination/lib/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
)

// Sample data for generating random employees
var (
	firstNames = []string{
		"Jan", "Piet", "Sophie", "Emma", "Lucas",
		"Daan", "Lotte", "Finn", "Julia", "Sem",
		"Anna", "Lars", "Sara", "Thomas", "Eva",
		"Tim", "Lisa", "Niels", "Fleur", "Ruben",
	}

	lastNames = []string{
		"de Vries", "Jansen", "van den Berg", "Bakker", "Visser",
		"Smit", "Meijer", "de Groot", "Mulder", "de Boer",
		"Vos", "Peters", "Hendriks", "van Leeuwen", "Dekker",
		"Brouwer", "de Wit", "Dijkstra", "Smits", "de Graaf",
	}

	roles = []string{
		"coordinator",
		"admin",
		"manager",
		"therapist",
		"nurse",
	}

	genders = []db.GenderEnum{
		db.GenderEnumMale,
		db.GenderEnumFemale,
		db.GenderEnumOther,
	}

	// Referring Organizations sample data
	orgNames = []string{
		"GGZ Noord-Holland",
		"Zorginstelling De Brug",
		"Mentale Gezondheid Utrecht",
		"Huisartspraktijk Centrum",
		"Sociaal Wijkteam Amsterdam",
		"GGZ Friesland",
		"Verslavingszorg Nederland",
		"Praktijk voor Psychologie Leiden",
		"Jeugdzorg Brabant",
		"Maatschappelijk Werk Rotterdam",
		"GGZ Limburg",
		"CAD Nijmegen",
		"Arkin Amsterdam",
		"Parnassia Groep",
		"Lentis Groningen",
	}

	orgContactFirstNames = []string{
		"Marieke", "Pieter", "Annemarie", "Wouter", "Liesbeth",
		"Jeroen", "Sandra", "Bas", "Ingrid", "Marcel",
	}

	// Care types for registration forms (excluding ambulatory_care which requires ambulatory_weekly_hours)
	careTypes = []db.CareTypeEnum{
		db.CareTypeEnumProtectedLiving,
		db.CareTypeEnumSemiIndependentLiving,
		db.CareTypeEnumIndependentAssistedLiving,
		// Note: ambulatory_care excluded because CreateClient doesn't support ambulatory_weekly_hours
		// which is required by the chk_ambulatory_hours constraint
	}

	// Registration reasons (Dutch)
	registrationReasons = []string{
		"Cliënt heeft ondersteuning nodig bij dagelijkse activiteiten",
		"Verwijzing van huisarts vanwege psychische klachten",
		"Behoefte aan beschermd wonen na ziekenhuisopname",
		"Cliënt kan niet meer zelfstandig wonen",
		"Verslavingsproblematiek vereist begeleide woonvorm",
		"Sociale isolatie en behoefte aan 24-uurs zorg",
		"Uitstroom uit GGZ-instelling, overgang naar begeleid wonen",
		"Jongvolwassene met ontwikkelingsstoornis zoekt passende woonvorm",
		"Mantelzorg valt weg, cliënt heeft externe hulp nodig",
		"Na scheiding tijdelijke woonplek met begeleiding nodig",
		"Cliënt herstelt van burn-out en heeft rust nodig",
		"Psychiatrische behandeling afgerond, begeleiding naar zelfstandigheid",
	}

	// Additional notes (Dutch)
	additionalNotes = []string{
		"Cliënt heeft voorkeur voor locatie in de buurt van familie",
		"Allergisch voor bepaalde medicatie",
		"Spreekt naast Nederlands ook Turks",
		"Heeft huisdier (kat) die mee moet kunnen",
		"Voorgeschiedenis van agressieve episode onder stress",
		"Werkt parttime en heeft flexibele begeleiding nodig",
		"Heeft jonge kinderen die op bezoek komen",
		"Rookt, heeft voorkeur voor rookruimte",
		"Vegetarisch dieet",
		"Heeft begeleiding bij financiën nodig",
		"",
		"",
		"",
	}

	// Location names (Dutch care facilities)
	locationNames = []string{
		"Woonlocatie De Zonnetuin",
		"Beschermd Wonen Centrum",
		"Huize De Linde",
		"Woongroep Horizon",
		"Villa Sereniteit",
		"Locatie Oost",
		"Huize Tranquil",
		"Wooncentrum De Haven",
		"Zorgvilla Het Park",
		"Locatie Nieuw Begin",
	}

	// Dutch cities for addresses
	dutchCities = []string{
		"Amsterdam", "Rotterdam", "Den Haag", "Utrecht", "Eindhoven",
		"Groningen", "Tilburg", "Almere", "Breda", "Nijmegen",
		"Haarlem", "Arnhem", "Zaanstad", "Amersfoort", "Apeldoorn",
	}

	// Street names
	streetNames = []string{
		"Hoofdstraat", "Kerkstraat", "Dorpsstraat", "Stationsweg",
		"Molenweg", "Julianalaan", "Wilhelminastraat", "Beatrixlaan",
		"Oranjestraat", "Zonnelaan", "Parklaan", "Bosweg",
	}

	// Family situations (Dutch)
	familySituations = []string{
		"Alleenstaand, geen kinderen",
		"Gescheiden, 2 kinderen (wonen bij ex-partner)",
		"Weduwe/weduwnaar, volwassen kinderen",
		"Gehuwd, partner woont thuis",
		"Alleenstaand, wekelijks contact met familie",
		"Ouders overleden, broer/zus als naaste familie",
		"Gescheiden, co-ouderschap met 1 kind",
		"Alleenstaand, beperkt contact met familie",
		"Samenlevend met partner, geen kinderen",
		"Familie woont in het buitenland",
	}

	// Limitations (Dutch)
	limitations = []string{
		"Beperkte mobiliteit, rollator nodig",
		"Verminderd gehoor, draagt gehoorapparaat",
		"Diabetespatiënt, insulineafhankelijk",
		"Cognitieve beperkingen door hersenletsel",
		"Angststoornis, heeft moeite met drukte",
		"Chronische vermoeidheid",
		"Beperkt zicht, leest met vergroting",
		"Geen fysieke beperkingen",
		"ADHD, heeft structuur nodig",
		"Autismespectrumstoornis",
	}

	// Focus areas (Dutch)
	focusAreas = []string{
		"Zelfstandig wonen en huishouden",
		"Sociale contacten opbouwen",
		"Dagstructuur en dagbesteding",
		"Financieel beheer",
		"Medicatietrouw",
		"Arbeidsparticipatie",
		"Omgaan met stress en emoties",
		"Verslavingsherstel",
		"Persoonlijke verzorging",
		"Opbouwen van een sociaal netwerk",
	}

	// Goals (Dutch)
	goals = []string{
		"Binnen 6 maanden zelfstandig boodschappen kunnen doen",
		"Stabiele dagstructuur ontwikkelen",
		"Werk of dagbesteding vinden binnen 3 maanden",
		"Medicatie zelfstandig kunnen beheren",
		"Schulden aflossen en financieel stabiel worden",
		"Sociale vaardigheden verbeteren",
		"Doorstromen naar zelfstandige woonvorm",
		"Clean blijven en terugval voorkomen",
		"Contact herstellen met familieleden",
		"Hobby of vrijetijdsbesteding vinden",
	}

	// Main providers (Dutch)
	mainProviders = []string{
		"GGZ Noord-Holland",
		"Gemeentelijke sociale dienst",
		"Huisarts",
		"Verslavingsarts",
		"Psychiater",
		"Maatschappelijk werker",
		"Ambulante begeleider",
		"Geen hoofdbehandelaar",
		"Praktijkondersteuner GGZ",
		"FACT-team",
	}

	// Intake notes (Dutch)
	intakeNotes = []string{
		"Cliënt is gemotiveerd om te werken aan doelen",
		"Kennismaking verliep positief, goede klik met team",
		"Cliënt toont inzicht in eigen situatie",
		"Afspraken gemaakt over huisregels en dagindeling",
		"Familie is betrokken bij intake",
		"Cliënt heeft vragen over uitkeringsaanvraag",
		"Medicatie moet nog worden afgestemd met behandelaar",
		"",
		"",
		"",
	}

	// Client notes (Dutch) - more detailed for ongoing care
	clientNotes = []string{
		"Cliënt maakt goede voortgang met dagstructuur",
		"Wekelijks contact met behandelaar gepland",
		"Medicatie is stabiel ingesteld",
		"Cliënt heeft moeite met groepsactiviteiten",
		"Goede samenwerking met externe hulpverleners",
		"Financiële situatie is verbeterd",
		"Cliënt werkt toe naar meer zelfstandigheid",
		"Aandachtspunt: sociale contacten uitbreiden",
		"",
		"",
	}

	// Waiting list priorities
	waitingListPriorities = []db.WaitingListPriorityEnum{
		db.WaitingListPriorityEnumLow,
		db.WaitingListPriorityEnumNormal,
		db.WaitingListPriorityEnumNormal,
		db.WaitingListPriorityEnumNormal,
		db.WaitingListPriorityEnumHigh,
	}

	// Closing reports (Dutch) - for discharged clients
	closingReports = []string{
		"Cliënt heeft alle behandeldoelen behaald en is klaar voor uitstroom naar zelfstandig wonen.",
		"Traject succesvol afgerond. Cliënt woont nu zelfstandig met ambulante begeleiding.",
		"Na 18 maanden is cliënt doorgestroomd naar beschermd wonen op eigen verzoek.",
		"Cliënt heeft zelf besloten te stoppen met begeleiding. Overdracht naar huisarts.",
		"Behandeling afgerond wegens verhuizing naar andere regio.",
		"Cliënt is hersteld en heeft geen verdere ondersteuning nodig.",
	}

	// Evaluation reports (Dutch) - for discharged clients
	evaluationReports = []string{
		"Gedurende het traject heeft cliënt grote stappen gemaakt in zelfstandigheid. Dagstructuur is verbeterd van chaotisch naar gestructureerd. Financiën worden nu zelfstandig beheerd.",
		"Cliënt kwam binnen met ernstige sociale angst. Na behandeling kan cliënt nu zelfstandig boodschappen doen en heeft een klein sociaal netwerk opgebouwd.",
		"Verslavingsproblematiek is onder controle. Cliënt is al 12 maanden clean en heeft werk gevonden.",
		"Psychiatrische symptomen zijn stabiel met huidige medicatie. Cliënt heeft inzicht in eigen situatie en weet wanneer hulp te zoeken.",
		"Traject was uitdagend vanwege terugval halverwege. Uiteindelijk toch positief afgerond dankzij intensieve begeleiding.",
		"Korter traject dan gepland, maar cliënt had sterke eigen motivatie en netwerk.",
	}

	// Discharge reasons
	dischargeReasons = []db.DischargeReasonEnum{
		db.DischargeReasonEnumTreatmentCompleted,
		db.DischargeReasonEnumTreatmentCompleted,
		db.DischargeReasonEnumTerminatedByMutualAgreement,
		db.DischargeReasonEnumTerminatedByClient,
		db.DischargeReasonEnumTerminatedDueToExternalFactors,
	}

	// Transfer reasons (Dutch) - for location transfers
	transferReasons = []string{
		"Cliënt heeft behoefte aan meer zelfstandigheid",
		"Betere match met zorgvraag op nieuwe locatie",
		"Dichter bij familie en sociaal netwerk",
		"Overplaatsing wegens capaciteitsproblemen",
		"Cliënt voelt zich niet thuis op huidige locatie",
		"Nieuwe coördinator met betere expertise voor cliënt",
		"Conflictsituatie op huidige locatie",
		"Medische reden: toegankelijkheid nieuwe locatie",
		"Wens van cliënt om dichter bij werk te wonen",
		"Afbouw naar lichtere zorgvorm",
	}

	// Incident types
	incidentTypes = []db.IncidentTypeEnum{
		db.IncidentTypeEnumAggression,
		db.IncidentTypeEnumMedicalEmergency,
		db.IncidentTypeEnumSafetyConcern,
		db.IncidentTypeEnumUnwantedBehavior,
		db.IncidentTypeEnumOther,
	}

	// Incident severities
	incidentSeverities = []db.IncidentSeverityEnum{
		db.IncidentSeverityEnumMinor,
		db.IncidentSeverityEnumMinor,
		db.IncidentSeverityEnumModerate,
		db.IncidentSeverityEnumModerate,
		db.IncidentSeverityEnumSevere,
	}

	// Incident statuses
	incidentStatuses = []db.IncidentStatusEnum{
		db.IncidentStatusEnumPending,
		db.IncidentStatusEnumUnderInvestigation,
		db.IncidentStatusEnumCompleted,
		db.IncidentStatusEnumCompleted,
	}

	// Incident descriptions (Dutch)
	incidentDescriptions = map[db.IncidentTypeEnum][]string{
		db.IncidentTypeEnumAggression: {
			"Cliënt werd verbaal agressief naar medebewoner tijdens avondeten",
			"Fysieke confrontatie tussen twee cliënten in gemeenschappelijke ruimte",
			"Cliënt schreeuwde en gooide voorwerpen in eigen kamer",
			"Verbale intimidatie richting begeleider tijdens medicatiemoment",
		},
		db.IncidentTypeEnumMedicalEmergency: {
			"Cliënt viel flauw in de gang, 112 gebeld",
			"Epileptische aanval tijdens groepsactiviteit",
			"Cliënt klaagde over pijn op de borst, ambulance ingeschakeld",
			"Allergische reactie na maaltijd, spoedhulp verleend",
		},
		db.IncidentTypeEnumSafetyConcern: {
			"Brandmelder ging af door roken op kamer",
			"Voordeur stond 's nachts open, cliënt was afwezig",
			"Gaslucht gemeld in keuken, noodprocedure gestart",
			"Cliënt uitgesproken gedachten over zelfbeschadiging",
		},
		db.IncidentTypeEnumUnwantedBehavior: {
			"Cliënt heeft alcohol gebruikt ondanks huisregels",
			"Bezoek na sluitingstijd zonder toestemming",
			"Cliënt weigerde deel te nemen aan verplichte groepsactiviteit",
			"Ongewenst bezoek van derden op locatie",
		},
		db.IncidentTypeEnumOther: {
			"Schade aan gemeenschappelijk meubilair",
			"Klacht van buren over geluidsoverlast",
			"Verlies van persoonlijke eigendommen gemeld",
			"Miscommunicatie over afspraken met externe hulpverlener",
		},
	}

	// Actions taken (Dutch)
	actionsTaken = map[db.IncidentTypeEnum][]string{
		db.IncidentTypeEnumAggression: {
			"Cliënten gescheiden, gesprek gevoerd met beide partijen, afkoelperiode ingesteld",
			"De-escalatie technieken toegepast, cliënt naar eigen kamer begeleid",
			"Incident besproken in teamoverleg, gedragsplan aangepast",
			"Waarschuwing gegeven, afspraken vastgelegd in begeleidingsplan",
		},
		db.IncidentTypeEnumMedicalEmergency: {
			"EHBO verleend, ambulance gebeld, cliënt naar ziekenhuis vervoerd",
			"Protocollen gevolgd, cliënt gestabiliseerd, arts geconsulteerd",
			"Spoedhulp verleend, medicatie toegediend, observatie gestart",
			"112 gebeld, eerste hulp verleend, familie geïnformeerd",
		},
		db.IncidentTypeEnumSafetyConcern: {
			"Ruimte geventileerd, brandweer geïnformeerd, cliënt aangesproken",
			"Cliënt teruggevonden, gesprek gevoerd, extra controles ingesteld",
			"Ontruimingsprotocol gestart, technische dienst ingeschakeld",
			"Direct gesprek met cliënt, crisisdienst ingeschakeld, 1-op-1 begeleiding",
		},
		db.IncidentTypeEnumUnwantedBehavior: {
			"Gesprek gevoerd over huisregels, consequenties besproken",
			"Bezoeker weggestuurd, afspraken aangescherpt",
			"Motiverend gesprek gehouden, alternatief aangeboden",
			"Huisregels herhaald, schriftelijke waarschuwing gegeven",
		},
		db.IncidentTypeEnumOther: {
			"Schade gedocumenteerd, herstelkosten berekend",
			"Excuses aangeboden aan buren, geluidsbeperkende maatregelen getroffen",
			"Zoektocht gestart, verzekering geïnformeerd",
			"Overleg met externe partij gepland, afspraken verduidelijkt",
		},
	}

	// Other parties (Dutch)
	otherParties = []string{
		"Medebewoner (naam geanonimiseerd)",
		"Familielid van cliënt",
		"Externe bezoeker",
		"Collega van andere afdeling",
		"Ambulancepersoneel",
		"Politie",
		"Buurman/buurvrouw",
		"",
		"",
		"",
	}
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	connPool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer connPool.Close()

	// Create store
	store := db.NewStore(connPool)

	// Seed locations first (needed for employees, intake forms, and clients)
	locationIDs, err := seedLocations(ctx, store, 8)
	if err != nil {
		log.Fatalf("Failed to seed locations: %v", err)
	}

	// Seed employees (returns employee IDs for coordinator assignment)
	employeeIDs, err := seedEmployees(ctx, store, 20, locationIDs)
	if err != nil {
		log.Fatalf("Failed to seed employees: %v", err)
	}

	// Seed referring organizations
	orgIDs, err := seedReferringOrganizations(ctx, store, 10)
	if err != nil {
		log.Fatalf("Failed to seed referring organizations: %v", err)
	}

	// Seed registration forms (returns form IDs for intake forms)
	// 30 total: 20 will have intakes, 10 will become waiting_list clients
	registrationFormIDs, err := seedRegistrationForms(ctx, store, orgIDs, 30)
	if err != nil {
		log.Fatalf("Failed to seed registration forms: %v", err)
	}

	// Seed intake forms for first 20 registration forms
	intakeInfos, err := seedIntakeForms(
		ctx,
		store,
		registrationFormIDs[:20],
		locationIDs,
		employeeIDs,
	)
	if err != nil {
		log.Fatalf("Failed to seed intake forms: %v", err)
	}

	// Seed clients with different statuses:
	// - 10 waiting_list clients (from registrations without intake)
	// - 8 in_care clients (from intake forms)
	// - 5 discharged clients (from intake forms)
	inCareClients, err := seedClients(
		ctx,
		store,
		registrationFormIDs[20:],
		intakeInfos,
		locationIDs,
		employeeIDs,
		orgIDs,
	)
	if err != nil {
		log.Fatalf("Failed to seed clients: %v", err)
	}

	// Seed location transfers for some in_care clients
	if err := seedLocationTransfers(ctx, store, inCareClients, locationIDs, employeeIDs); err != nil {
		log.Fatalf("Failed to seed location transfers: %v", err)
	}

	// Seed incidents for in_care clients
	if err := seedIncidents(ctx, store, inCareClients, locationIDs, employeeIDs); err != nil {
		log.Fatalf("Failed to seed incidents: %v", err)
	}

	// Seed evaluations for in_care clients
	if err := seedEvaluations(ctx, store, inCareClients); err != nil {
		log.Fatalf("Failed to seed evaluations: %v", err)
	}

	// Seed appointments for admin user
	if err := seedAppointments(ctx, store, cfg.AdminEmail, inCareClients); err != nil {
		log.Fatalf("Failed to seed appointments: %v", err)
	}

	// Seed notifications for admin users
	if err := seedNotifications(ctx, store); err != nil {
		log.Fatalf("Failed to seed notifications: %v", err)
	}

	// Seed audit logs for realistic activity tracking
	// Get users and clients from seeded data for realistic audit entries
	adminUserIDs, _ := store.GetUserIDsByRoleName(ctx, "admin")
	coordinatorUserIDs, _ := store.GetUserIDsByRoleName(ctx, "coordinator")
	allUserIDs := append(adminUserIDs, coordinatorUserIDs...)

	if len(allUserIDs) == 0 {
		fmt.Println("  ⚠ No users found, skipping audit log seeding")
	} else {
		// Get all clients for client-related audit entries
		inCareClients, _ := store.ListInCareClients(ctx, db.ListInCareClientsParams{Limit: 1000, Offset: 0})
		waitingListClients, _ := store.ListWaitingListClients(ctx, db.ListWaitingListClientsParams{Limit: 1000, Offset: 0})
		var clientIDs []string
		for _, c := range inCareClients {
			clientIDs = append(clientIDs, c.ID)
		}
		for _, c := range waitingListClients {
			clientIDs = append(clientIDs, c.ID)
		}

		// Get all employees for employee-related audit entries
		employees, _ := store.ListEmployees(ctx, db.ListEmployeesParams{Limit: 100, Offset: 0})
		var employeeIDs []string
		for _, e := range employees {
			employeeIDs = append(employeeIDs, e.ID)
		}

		// Create 500 audit logs spread over the last 30 days
		if err := seedAuditLogs(ctx, store, allUserIDs, employeeIDs, clientIDs, 30, 500); err != nil {
			log.Fatalf("Failed to seed audit logs: %v", err)
		}
	}

	fmt.Println("✅ Successfully seeded database!")
}

func seedEmployees(ctx context.Context, store *db.Store, count int, locationIDs []string) ([]string, error) {
	fmt.Printf("🌱 Seeding %d employees...\n", count)

	employeeIDs := make([]string, 0, count)

	for i := 0; i < count; i++ {
		employee, err := createRandomEmployee(ctx, store, locationIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to create employee %d: %w", i+1, err)
		}
		employeeIDs = append(employeeIDs, employee.ID)
		fmt.Printf(
			"  ✓ Created employee: %s %s (%s)\n",
			employee.FirstName,
			employee.LastName,
			employee.Role,
		)
	}

	fmt.Printf("✅ Successfully seeded %d employees\n", count)
	return employeeIDs, nil
}

type EmployeeInfo struct {
	ID        string
	FirstName string
	LastName  string
	Role      string
}

func createRandomEmployee(ctx context.Context, store *db.Store, locationIDs []string) (*EmployeeInfo, error) {
	// Generate random data
	firstName := randomElement(firstNames)
	lastName := randomElement(lastNames)
	email := generateEmail(firstName, lastName)
	role := randomElement(roles)
	locationID := randomElement(locationIDs)

	// Generate IDs
	userID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}
	employeeID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate employee ID: %w", err)
	}

	// Hash a default password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the employee using the transaction
	err = store.CreateEmployeeTx(ctx, db.CreateEmployeeTxParams{
		User: db.CreateUserParams{
			ID:           userID,
			Email:        email,
			PasswordHash: string(passwordHash),
		},
		Emp: db.CreateEmployeeParams{
			ID:          employeeID,
			UserID:      userID, // Will be overwritten in tx, but need to provide
			FirstName:   firstName,
			LastName:    lastName,
			Bsn:         generateBSN(),
			DateOfBirth: generateRandomDateOfBirth(),
			PhoneNumber: generatePhoneNumber(),
			Gender:      randomGender(),
			LocationID:  locationID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create employee tx: %w", err)
	}

	return &EmployeeInfo{
		ID:        employeeID,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
	}, nil
}

// Helper functions

func randomElement[T any](slice []T) T {
	return slice[rand.Intn(len(slice))]
}

func randomGender() db.GenderEnum {
	return genders[rand.Intn(len(genders))]
}

func generateEmail(firstName, lastName string) string {
	// Create a unique email using nanoid suffix
	suffix, _ := gonanoid.Generate("0123456789", 4)
	return fmt.Sprintf("%s.%s.%s@example.com",
		normalizeForEmail(firstName),
		normalizeForEmail(lastName),
		suffix)
}

func normalizeForEmail(s string) string {
	// Simple normalization - lowercase and replace spaces
	result := ""
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			result += string(c)
		} else if c >= 'A' && c <= 'Z' {
			result += string(c + 32) // Convert to lowercase
		}
	}
	return result
}

func generateBSN() string {
	// Generate a random 9-digit BSN (Dutch social security number)
	// Note: This doesn't follow the actual BSN validation rules, just random digits
	return fmt.Sprintf("%09d", rand.Intn(1000000000))
}

func generateRandomDateOfBirth() pgtype.Date {
	// Generate a date of birth between 20 and 60 years ago
	minAge, maxAge := 20, 60
	yearsAgo := minAge + rand.Intn(maxAge-minAge)

	dob := time.Now().AddDate(-yearsAgo, -rand.Intn(12), -rand.Intn(28))

	return pgtype.Date{
		Time:  dob,
		Valid: true,
	}
}

func generatePhoneNumber() string {
	// Generate a Dutch mobile phone number format
	return fmt.Sprintf("06%08d", rand.Intn(100000000))
}

func generateLandlineNumber() string {
	// Generate a Dutch landline phone number format (area codes 010-079)
	areaCode := 10 + rand.Intn(70)
	return fmt.Sprintf("0%d-%07d", areaCode, rand.Intn(10000000))
}

// ============================================================
// Referring Organizations Seeding
// ============================================================

func seedReferringOrganizations(ctx context.Context, store *db.Store, count int) ([]string, error) {
	fmt.Printf("🌱 Seeding %d referring organizations...\n", count)

	orgIDs := make([]string, 0, count)

	for i := 0; i < count && i < len(orgNames); i++ {
		org, err := createRandomReferringOrg(ctx, store, i)
		if err != nil {
			return nil, fmt.Errorf("failed to create referring org %d: %w", i+1, err)
		}
		orgIDs = append(orgIDs, org.ID)
		fmt.Printf("  ✓ Created referring org: %s (%s)\n", org.Name, org.ContactPerson)
	}

	fmt.Printf("✅ Successfully seeded %d referring organizations\n", len(orgIDs))
	return orgIDs, nil
}

type ReferringOrgInfo struct {
	ID            string
	Name          string
	ContactPerson string
}

func createRandomReferringOrg(
	ctx context.Context,
	store *db.Store,
	index int,
) (*ReferringOrgInfo, error) {
	// Generate ID
	orgID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate org ID: %w", err)
	}

	// Use predefined org name or generate one
	orgName := orgNames[index%len(orgNames)]

	// Generate contact person
	contactFirstName := randomElement(orgContactFirstNames)
	contactLastName := randomElement(lastNames)
	contactPerson := fmt.Sprintf("%s %s", contactFirstName, contactLastName)

	// Generate email based on org name
	suffix, _ := gonanoid.Generate("0123456789", 3)
	orgEmail := fmt.Sprintf("info.%s@%s.nl",
		suffix,
		normalizeForEmail(orgName))

	err = store.CreateReferringOrg(ctx, db.CreateReferringOrgParams{
		ID:            orgID,
		Name:          orgName,
		ContactPerson: contactPerson,
		PhoneNumber:   generateLandlineNumber(),
		Email:         orgEmail,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create referring org: %w", err)
	}

	return &ReferringOrgInfo{
		ID:            orgID,
		Name:          orgName,
		ContactPerson: contactPerson,
	}, nil
}

// ============================================================
// Registration Forms Seeding
// ============================================================

func seedRegistrationForms(
	ctx context.Context,
	store *db.Store,
	orgIDs []string,
	count int,
) ([]string, error) {
	fmt.Printf("🌱 Seeding %d registration forms...\n", count)

	formIDs := make([]string, 0, count)

	for i := 0; i < count; i++ {
		form, err := createRandomRegistrationForm(ctx, store, orgIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to create registration form %d: %w", i+1, err)
		}
		formIDs = append(formIDs, form.ID)
		fmt.Printf("  ✓ Created registration: %s %s (BSN: %s)\n",
			form.FirstName, form.LastName, form.BSN)
	}

	fmt.Printf("✅ Successfully seeded %d registration forms\n", count)
	return formIDs, nil
}

type RegistrationFormInfo struct {
	ID        string
	FirstName string
	LastName  string
	BSN       string
}

func createRandomRegistrationForm(
	ctx context.Context,
	store *db.Store,
	orgIDs []string,
) (*RegistrationFormInfo, error) {
	// Generate ID
	formID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate form ID: %w", err)
	}

	// Generate random client data
	firstName := randomElement(firstNames)
	lastName := randomElement(lastNames)
	bsn := generateBSN()

	// Randomly assign a referring organization (80% chance) or nil (20% chance)
	var refferingOrgID *string
	if rand.Float32() < 0.8 && len(orgIDs) > 0 {
		orgID := randomElement(orgIDs)
		refferingOrgID = &orgID
	}

	// Generate date of birth for client (18-75 years old)
	dob := generateClientDateOfBirth()

	// Generate registration date (within last 90 days)
	regDate := generateRecentDate(90)

	// Random care type
	careType := randomElement(careTypes)

	// Random registration reason
	reason := randomElement(registrationReasons)

	// Random additional notes (may be nil)
	var notes *string
	note := randomElement(additionalNotes)
	if note != "" {
		notes = &note
	}

	err = store.CreateRegistrationForm(ctx, db.CreateRegistrationFormParams{
		ID:                 formID,
		FirstName:          firstName,
		LastName:           lastName,
		Bsn:                bsn,
		Gender:             randomGender(),
		DateOfBirth:        dob,
		RefferingOrgID:     refferingOrgID,
		CareType:           careType,
		RegistrationDate:   regDate,
		RegistrationReason: reason,
		AdditionalNotes:    notes,
		AttachmentIds:      []string{}, // Empty for seeding
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create registration form: %w", err)
	}

	return &RegistrationFormInfo{
		ID:        formID,
		FirstName: firstName,
		LastName:  lastName,
		BSN:       bsn,
	}, nil
}

func generateClientDateOfBirth() pgtype.Date {
	// Generate a date of birth between 18 and 75 years ago
	minAge, maxAge := 18, 75
	yearsAgo := minAge + rand.Intn(maxAge-minAge)

	dob := time.Now().AddDate(-yearsAgo, -rand.Intn(12), -rand.Intn(28))

	return pgtype.Date{
		Time:  dob,
		Valid: true,
	}
}

func generateRecentDate(daysAgo int) pgtype.Date {
	// Generate a random date within the last N days
	days := rand.Intn(daysAgo)
	date := time.Now().AddDate(0, 0, -days)

	return pgtype.Date{
		Time:  date,
		Valid: true,
	}
}

// ============================================================
// Locations Seeding
// ============================================================

func seedLocations(ctx context.Context, store *db.Store, count int) ([]string, error) {
	fmt.Printf("🌱 Seeding %d locations...\n", count)

	locationIDs := make([]string, 0, count)

	for i := 0; i < count && i < len(locationNames); i++ {
		loc, err := createRandomLocation(ctx, store, i)
		if err != nil {
			return nil, fmt.Errorf("failed to create location %d: %w", i+1, err)
		}
		locationIDs = append(locationIDs, loc.ID)
		fmt.Printf("  ✓ Created location: %s (capacity: %d)\n", loc.Name, loc.Capacity)
	}

	fmt.Printf("✅ Successfully seeded %d locations\n", len(locationIDs))
	return locationIDs, nil
}

type LocationInfo struct {
	ID       string
	Name     string
	Capacity int32
}

func createRandomLocation(ctx context.Context, store *db.Store, index int) (*LocationInfo, error) {
	// Generate ID
	locID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate location ID: %w", err)
	}

	// Use predefined location name
	locName := locationNames[index%len(locationNames)]

	// Generate address
	streetName := randomElement(streetNames)
	houseNumber := rand.Intn(200) + 1
	city := randomElement(dutchCities)
	address := fmt.Sprintf("%s %d, %s", streetName, houseNumber, city)

	// Generate postal code (Dutch format: 4 digits + 2 letters)
	postalCode := fmt.Sprintf("%04d%s", 1000+rand.Intn(9000), randomPostalLetters())

	// Random capacity between 8 and 30
	capacity := int32(8 + rand.Intn(23))

	// Random occupied (0 to capacity-1, leave some room)
	occupied := int32(rand.Intn(int(capacity)))

	err = store.CreateLocation(ctx, db.CreateLocationParams{
		ID:         locID,
		Name:       locName,
		PostalCode: postalCode,
		Address:    address,
		Capacity:   capacity,
		Occupied:   occupied,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	return &LocationInfo{
		ID:       locID,
		Name:     locName,
		Capacity: capacity,
	}, nil
}

func randomPostalLetters() string {
	letters := "ABCDEFGHJKLMNPRSTUVWXYZ" // Dutch postal codes don't use all letters
	return string(letters[rand.Intn(len(letters))]) + string(letters[rand.Intn(len(letters))])
}

// ============================================================
// Intake Forms Seeding
// ============================================================

func seedIntakeForms(
	ctx context.Context,
	store *db.Store,
	registrationFormIDs, locationIDs, employeeIDs []string,
) ([]IntakeFormInfo, error) {
	count := len(registrationFormIDs)
	fmt.Printf("🌱 Seeding %d intake forms...\n", count)

	intakeInfos := make([]IntakeFormInfo, 0, count)

	for i, regFormID := range registrationFormIDs {
		intake, err := createRandomIntakeForm(ctx, store, regFormID, locationIDs, employeeIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to create intake form %d: %w", i+1, err)
		}
		intakeInfos = append(intakeInfos, *intake)
		fmt.Printf("  ✓ Created intake: %s (location: %s)\n", intake.ID[:8], intake.LocationID[:8])
	}

	fmt.Printf("✅ Successfully seeded %d intake forms\n", count)
	return intakeInfos, nil
}

type IntakeFormInfo struct {
	ID                 string
	RegistrationFormID string
	LocationID         string
	CoordinatorID      string
	FamilySituation    *string
	Limitations        *string
	FocusAreas         *string
	Goals              []string
	Notes              *string
}

func createRandomIntakeForm(
	ctx context.Context,
	store *db.Store,
	registrationFormID string,
	locationIDs, employeeIDs []string,
) (*IntakeFormInfo, error) {
	// Generate ID
	intakeID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate intake ID: %w", err)
	}

	// Pick random location and coordinator
	locationID := randomElement(locationIDs)
	coordinatorID := randomElement(employeeIDs)

	// Generate intake date (within last 60 days)
	intakeDate := generateRecentDate(60)

	// Generate intake time (between 9:00 and 17:00)
	hour := 9 + rand.Intn(8)
	minute := rand.Intn(4) * 15 // 0, 15, 30, or 45
	intakeTime := pgtype.Time{
		Microseconds: int64(hour*3600+minute*60) * 1000000,
		Valid:        true,
	}

	// Random optional fields
	var familySituation, mainProvider, limitationsStr, focusAreasStr, notes *string
	var goalsStr []string

	if rand.Float32() < 0.9 {
		fs := randomElement(familySituations)
		familySituation = &fs
	}
	if rand.Float32() < 0.8 {
		mp := randomElement(mainProviders)
		mainProvider = &mp
	}
	if rand.Float32() < 0.85 {
		l := randomElement(limitations)
		limitationsStr = &l
	}
	if rand.Float32() < 0.9 {
		fa := randomElement(focusAreas)
		focusAreasStr = &fa
	}
	if rand.Float32() < 0.9 {
		g := randomElement(goals)
		goalsStr = []string{g}
	}
	// Goals: convert strings to params
	seededGoals := make([]db.CreateClientGoalParams, len(goalsStr))
	for i, g := range goalsStr {
		gid, _ := gonanoid.New()
		seededGoals[i] = db.CreateClientGoalParams{
			ID:           gid,
			IntakeFormID: intakeID,
			Title:        g,
		}
	}

	// Notes: include if not empty string
	noteContent := randomElement(intakeNotes)
	if noteContent != "" {
		notes = &noteContent
	}

	// Evaluation Interval (4-12 weeks)
	interval := int32(4 + rand.Intn(9))

	_, err = store.CreateIntakeFormTx(ctx, db.CreateIntakeFormTxParams{
		IntakeForm: db.CreateIntakeFormParams{
			ID:                      intakeID,
			RegistrationFormID:      registrationFormID,
			IntakeDate:              intakeDate,
			IntakeTime:              intakeTime,
			LocationID:              locationID,
			CoordinatorID:           coordinatorID,
			FamilySituation:         familySituation,
			MainProvider:            mainProvider,
			Limitations:             limitationsStr,
			FocusAreas:              focusAreasStr,
			Notes:                   notes,
			EvaluationIntervalWeeks: &interval,
		},
		RegistrationFormID: registrationFormID,
		RegistrationFormStatus: db.NullRegistrationStatusEnum{
			RegistrationStatusEnum: db.RegistrationStatusEnumApproved,
			Valid:                  true,
		},
		Goals: seededGoals,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create intake form: %w", err)
	}

	return &IntakeFormInfo{
		ID:                 intakeID,
		RegistrationFormID: registrationFormID,
		LocationID:         locationID,
		CoordinatorID:      coordinatorID,
		FamilySituation:    familySituation,
		Limitations:        limitationsStr,
		FocusAreas:         focusAreasStr,
		Goals:              goalsStr,
		Notes:              notes,
	}, nil
}

// ============================================================
// Clients Seeding
// ============================================================

// ClientInfo holds info about created clients for use in subsequent seeding
type ClientInfo struct {
	ID            string
	LocationID    string
	CoordinatorID string
}

func seedClients(
	ctx context.Context,
	store *db.Store,
	waitingListRegIDs []string,
	intakeInfos []IntakeFormInfo,
	locationIDs, employeeIDs, orgIDs []string,
) ([]ClientInfo, error) {
	fmt.Println("🌱 Seeding clients...")

	// Get registration forms data for waiting list clients
	waitingListCount := len(waitingListRegIDs)
	inCareCount := 8
	dischargedCount := 5

	if inCareCount+dischargedCount > len(intakeInfos) {
		inCareCount = len(intakeInfos) - dischargedCount
		if inCareCount < 0 {
			inCareCount = 0
			dischargedCount = len(intakeInfos)
		}
	}

	inCareClients := make([]ClientInfo, 0, inCareCount)

	// Seed waiting list clients (from registrations without intake)
	fmt.Printf("  Seeding %d waiting_list clients...\n", waitingListCount)
	for i, regFormID := range waitingListRegIDs {
		err := createWaitingListClient(ctx, store, regFormID, locationIDs, employeeIDs, orgIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to create waiting_list client %d: %w", i+1, err)
		}
	}
	fmt.Printf("  ✓ Created %d waiting_list clients\n", waitingListCount)

	// Seed in_care clients (from intake forms)
	fmt.Printf("  Seeding %d in_care clients...\n", inCareCount)
	for i := 0; i < inCareCount; i++ {
		clientInfo, err := createInCareClient(ctx, store, intakeInfos[i], orgIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to create in_care client %d: %w", i+1, err)
		}
		inCareClients = append(inCareClients, *clientInfo)
	}
	fmt.Printf("  ✓ Created %d in_care clients\n", inCareCount)

	// Seed discharged clients (from remaining intake forms)
	fmt.Printf("  Seeding %d discharged clients...\n", dischargedCount)
	for i := 0; i < dischargedCount; i++ {
		intakeIdx := inCareCount + i
		err := createDischargedClient(ctx, store, intakeInfos[intakeIdx], orgIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to create discharged client %d: %w", i+1, err)
		}
	}
	fmt.Printf("  ✓ Created %d discharged clients\n", dischargedCount)

	totalClients := waitingListCount + inCareCount + dischargedCount
	fmt.Printf("✅ Successfully seeded %d clients\n", totalClients)
	return inCareClients, nil
}

func createWaitingListClient(
	ctx context.Context,
	store *db.Store,
	registrationFormID string,
	locationIDs, employeeIDs, orgIDs []string,
) error {
	// Get registration form data
	regForm, err := store.GetRegistrationForm(ctx, registrationFormID)
	if err != nil {
		return fmt.Errorf("failed to get registration form: %w", err)
	}

	clientID, err := gonanoid.New()
	if err != nil {
		return fmt.Errorf("failed to generate client ID: %w", err)
	}

	// For waiting list, we need a temporary intake form ID (create a placeholder)
	intakeID, err := gonanoid.New()
	if err != nil {
		return fmt.Errorf("failed to generate intake ID: %w", err)
	}

	// Create placeholder intake form for waiting list client
	locationID := randomElement(locationIDs)
	coordinatorID := randomElement(employeeIDs)

	_, err = store.CreateIntakeFormTx(ctx, db.CreateIntakeFormTxParams{
		IntakeForm: db.CreateIntakeFormParams{
			ID:                 intakeID,
			RegistrationFormID: registrationFormID,
			IntakeDate:         generateRecentDate(30),
			IntakeTime: pgtype.Time{
				Microseconds: int64(10*3600) * 1000000,
				Valid:        true,
			},
			LocationID:    locationID,
			CoordinatorID: coordinatorID,
		},
		RegistrationFormID: registrationFormID,
		RegistrationFormStatus: db.NullRegistrationStatusEnum{
			RegistrationStatusEnum: db.RegistrationStatusEnumApproved,
			Valid:                  true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create placeholder intake: %w", err)
	}

	// Random phone number
	phone := generatePhoneNumber()

	// Random focus areas and notes for waiting list
	var focusAreasStr, notesStr *string
	if rand.Float32() < 0.8 {
		fa := randomElement(focusAreas)
		focusAreasStr = &fa
	}
	noteContent := randomElement(clientNotes)
	if noteContent != "" {
		notesStr = &noteContent
	}

	// Pick random referring org (may be same as registration form's org or different)
	var referringOrgID *string
	if regForm.RefferingOrgID != nil {
		referringOrgID = regForm.RefferingOrgID
	} else if rand.Float32() < 0.7 {
		orgID := randomElement(orgIDs)
		referringOrgID = &orgID
	}

	_, err = store.CreateClient(ctx, db.CreateClientParams{
		ID:                  clientID,
		FirstName:           regForm.FirstName,
		LastName:            regForm.LastName,
		Bsn:                 regForm.Bsn,
		DateOfBirth:         regForm.DateOfBirth,
		PhoneNumber:         &phone,
		Gender:              regForm.Gender,
		RegistrationFormID:  registrationFormID,
		IntakeFormID:        intakeID,
		CareType:            regForm.CareType,
		ReferringOrgID:      referringOrgID,
		WaitingListPriority: randomElement(waitingListPriorities),
		Status:              db.ClientStatusEnumWaitingList,
		AssignedLocationID:  locationID,
		CoordinatorID:       coordinatorID,
		FocusAreas:          focusAreasStr,
		Notes:               notesStr,
	})
	if err != nil {
		return err
	}

	// Link goals to the client (update client_id in goals table)
	err = store.LinkGoalsToClient(ctx, db.LinkGoalsToClientParams{
		ClientID:     &clientID,
		IntakeFormID: intakeID,
	})
	if err != nil {
		fmt.Printf("  ⚠ Warning: Failed to link goals for client %s: %v\n", clientID, err)
	}

	// Update intake form status to completed since client is now in waiting list
	return store.UpdateIntakeFormStatus(ctx, db.UpdateIntakeFormStatusParams{
		ID:     intakeID,
		Status: db.IntakeStatusEnumCompleted,
	})
}

func createInCareClient(
	ctx context.Context,
	store *db.Store,
	intakeInfo IntakeFormInfo,
	orgIDs []string,
) (*ClientInfo, error) {
	// Get registration form data
	regForm, err := store.GetRegistrationForm(ctx, intakeInfo.RegistrationFormID)
	if err != nil {
		return nil, fmt.Errorf("failed to get registration form: %w", err)
	}

	clientID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client ID: %w", err)
	}

	// Random phone number
	phone := generatePhoneNumber()

	// Pick random referring org
	var referringOrgID *string
	if regForm.RefferingOrgID != nil {
		referringOrgID = regForm.RefferingOrgID
	} else if rand.Float32() < 0.7 {
		orgID := randomElement(orgIDs)
		referringOrgID = &orgID
	}

	// Generate care start date (30-180 days ago for in-care)
	careStartDate := generateRecentDate(180)

	// Client notes for in-care clients
	var notesStr *string
	noteContent := randomElement(clientNotes)
	if noteContent != "" {
		notesStr = &noteContent
	}

	// Create client first as waiting_list (to satisfy constraints)
	_, err = store.CreateClient(ctx, db.CreateClientParams{
		ID:                  clientID,
		FirstName:           regForm.FirstName,
		LastName:            regForm.LastName,
		Bsn:                 regForm.Bsn,
		DateOfBirth:         regForm.DateOfBirth,
		PhoneNumber:         &phone,
		Gender:              regForm.Gender,
		RegistrationFormID:  intakeInfo.RegistrationFormID,
		IntakeFormID:        intakeInfo.ID,
		CareType:            regForm.CareType,
		ReferringOrgID:      referringOrgID,
		WaitingListPriority: db.WaitingListPriorityEnumNormal,
		Status:              db.ClientStatusEnumWaitingList, // Create as waiting_list first
		AssignedLocationID:  intakeInfo.LocationID,
		CoordinatorID:       intakeInfo.CoordinatorID,
		FamilySituation:     intakeInfo.FamilySituation,
		Limitations:         intakeInfo.Limitations,
		FocusAreas:          intakeInfo.FocusAreas,
		Notes:               notesStr,
	})
	if err != nil {
		return nil, err
	}

	// Update to in_care with care_start_date in single call to satisfy chk_in_care_fields constraint
	_, err = store.UpdateClient(ctx, db.UpdateClientParams{
		ID: clientID,
		Status: db.NullClientStatusEnum{
			ClientStatusEnum: db.ClientStatusEnumInCare,
			Valid:            true,
		},
		CareStartDate: careStartDate,
	})
	if err != nil {
		return nil, err
	}

	// Link goals to the client (update client_id in goals table)
	err = store.LinkGoalsToClient(ctx, db.LinkGoalsToClientParams{
		ClientID:     &clientID,
		IntakeFormID: intakeInfo.ID,
	})
	if err != nil {
		fmt.Printf("  ⚠ Warning: Failed to link goals for client %s: %v\n", clientID, err)
	}

	// Update intake form status to completed since client is now in care
	err = store.UpdateIntakeFormStatus(ctx, db.UpdateIntakeFormStatusParams{
		ID:     intakeInfo.ID,
		Status: db.IntakeStatusEnumCompleted,
	})
	if err != nil {
		return nil, err
	}

	return &ClientInfo{
		ID:            clientID,
		LocationID:    intakeInfo.LocationID,
		CoordinatorID: intakeInfo.CoordinatorID,
	}, nil
}

func createDischargedClient(
	ctx context.Context,
	store *db.Store,
	intakeInfo IntakeFormInfo,
	orgIDs []string,
) error {
	// Get registration form data
	regForm, err := store.GetRegistrationForm(ctx, intakeInfo.RegistrationFormID)
	if err != nil {
		return fmt.Errorf("failed to get registration form: %w", err)
	}

	clientID, err := gonanoid.New()
	if err != nil {
		return fmt.Errorf("failed to generate client ID: %w", err)
	}

	// Random phone number
	phone := generatePhoneNumber()

	// Pick random referring org
	var referringOrgID *string
	if regForm.RefferingOrgID != nil {
		referringOrgID = regForm.RefferingOrgID
	} else if rand.Float32() < 0.7 {
		orgID := randomElement(orgIDs)
		referringOrgID = &orgID
	}

	// Generate dates for discharged clients
	// Care started 6-18 months ago
	careStartDaysAgo := 180 + rand.Intn(365)
	careStartDate := pgtype.Date{
		Time:  time.Now().AddDate(0, 0, -careStartDaysAgo),
		Valid: true,
	}

	// Discharged 7-60 days ago
	dischargeDaysAgo := 7 + rand.Intn(53)
	dischargeDate := pgtype.Date{
		Time:  time.Now().AddDate(0, 0, -dischargeDaysAgo),
		Valid: true,
	}

	// Discharge-specific data
	closingReport := randomElement(closingReports)
	evaluationReport := randomElement(evaluationReports)
	dischargeReason := randomElement(dischargeReasons)

	// STEP 1: Create client as waiting_list (initial state)
	_, err = store.CreateClient(ctx, db.CreateClientParams{
		ID:                  clientID,
		FirstName:           regForm.FirstName,
		LastName:            regForm.LastName,
		Bsn:                 regForm.Bsn,
		DateOfBirth:         regForm.DateOfBirth,
		PhoneNumber:         &phone,
		Gender:              regForm.Gender,
		RegistrationFormID:  intakeInfo.RegistrationFormID,
		IntakeFormID:        intakeInfo.ID,
		CareType:            regForm.CareType,
		ReferringOrgID:      referringOrgID,
		WaitingListPriority: db.WaitingListPriorityEnumNormal,
		Status:              db.ClientStatusEnumWaitingList,
		AssignedLocationID:  intakeInfo.LocationID,
		CoordinatorID:       intakeInfo.CoordinatorID,
		FamilySituation:     intakeInfo.FamilySituation,
		Limitations:         intakeInfo.Limitations,
		FocusAreas:          intakeInfo.FocusAreas,
	})
	if err != nil {
		return fmt.Errorf("step 1 (create waiting_list) failed: %w", err)
	}

	// STEP 2: Move to in_care with care_start_date
	_, err = store.UpdateClient(ctx, db.UpdateClientParams{
		ID: clientID,
		Status: db.NullClientStatusEnum{
			ClientStatusEnum: db.ClientStatusEnumInCare,
			Valid:            true,
		},
		CareStartDate: careStartDate,
	})
	if err != nil {
		return fmt.Errorf("step 2 (move to in_care) failed: %w", err)
	}

	// STEP 3: Discharge client with all required discharge fields
	_, err = store.UpdateClient(ctx, db.UpdateClientParams{
		ID: clientID,
		Status: db.NullClientStatusEnum{
			ClientStatusEnum: db.ClientStatusEnumDischarged,
			Valid:            true,
		},
		DischargeDate:    dischargeDate,
		ClosingReport:    &closingReport,
		EvaluationReport: &evaluationReport,
		ReasonForDischarge: db.NullDischargeReasonEnum{
			DischargeReasonEnum: dischargeReason,
			Valid:               true,
		},
		DischargeStatus: db.NullDischargeStatusEnum{
			DischargeStatusEnum: db.DischargeStatusEnumCompleted,
			Valid:               true,
		},
	})
	if err != nil {
		return fmt.Errorf("step 3 (discharge) failed: %w", err)
	}

	// Link goals to the client (update client_id in goals table)
	err = store.LinkGoalsToClient(ctx, db.LinkGoalsToClientParams{
		ClientID:     &clientID,
		IntakeFormID: intakeInfo.ID,
	})
	if err != nil {
		fmt.Printf("  ⚠ Warning: Failed to link goals for client %s: %v\n", clientID, err)
	}

	// Update intake form status to completed since client was processed
	err = store.UpdateIntakeFormStatus(ctx, db.UpdateIntakeFormStatusParams{
		ID:     intakeInfo.ID,
		Status: db.IntakeStatusEnumCompleted,
	})
	if err != nil {
		return fmt.Errorf("step 4 (update intake status) failed: %w", err)
	}

	return nil
}

// ============================================================
// Location Transfers Seeding
// ============================================================

func seedLocationTransfers(
	ctx context.Context,
	store *db.Store,
	inCareClients []ClientInfo,
	locationIDs, employeeIDs []string,
) error {
	// Create transfers for about half of the in_care clients
	transferCount := len(inCareClients) / 2
	if transferCount < 1 && len(inCareClients) > 0 {
		transferCount = 1
	}

	fmt.Printf("🌱 Seeding %d location transfers...\n", transferCount)

	for i := 0; i < transferCount; i++ {
		client := inCareClients[i]

		// Find a different location for the transfer
		var newLocationID string
		for _, locID := range locationIDs {
			if locID != client.LocationID {
				newLocationID = locID
				break
			}
		}
		// If all locations are the same, just use the first different one or same
		if newLocationID == "" {
			newLocationID = locationIDs[0]
		}

		// Find a different coordinator
		var newCoordinatorID string
		for _, empID := range employeeIDs {
			if empID != client.CoordinatorID {
				newCoordinatorID = empID
				break
			}
		}
		if newCoordinatorID == "" {
			newCoordinatorID = employeeIDs[0]
		}

		err := createLocationTransfer(ctx, store, client, newLocationID, newCoordinatorID)
		if err != nil {
			return fmt.Errorf("failed to create location transfer %d: %w", i+1, err)
		}
	}

	fmt.Printf("✅ Successfully seeded %d location transfers\n", transferCount)
	return nil
}

func createLocationTransfer(
	ctx context.Context,
	store *db.Store,
	client ClientInfo,
	newLocationID, newCoordinatorID string,
) error {
	transferID, err := gonanoid.New()
	if err != nil {
		return fmt.Errorf("failed to generate transfer ID: %w", err)
	}

	// Generate a transfer date (within last 30 days)
	daysAgo := rand.Intn(30)
	transferDate := pgtype.Timestamp{
		Time:  time.Now().AddDate(0, 0, -daysAgo),
		Valid: true,
	}

	// Random transfer reason
	reason := randomElement(transferReasons)

	_, err = store.CreateLocationTransfer(ctx, db.CreateLocationTransferParams{
		ID:                   transferID,
		ClientID:             client.ID,
		FromLocationID:       &client.LocationID,
		ToLocationID:         newLocationID,
		CurrentCoordinatorID: client.CoordinatorID,
		NewCoordinatorID:     newCoordinatorID,
		TransferDate:         transferDate,
		Reason:               &reason,
	})

	return err
}

// ============================================================
// Incidents Seeding
// ============================================================

func seedIncidents(
	ctx context.Context,
	store *db.Store,
	inCareClients []ClientInfo,
	locationIDs, employeeIDs []string,
) error {
	// Create 2-4 incidents per in_care client
	fmt.Println("🌱 Seeding incidents...")

	totalIncidents := 0
	for _, client := range inCareClients {
		// Random number of incidents per client (1-3)
		numIncidents := 1 + rand.Intn(3)
		for j := 0; j < numIncidents; j++ {
			err := createIncident(ctx, store, client)
			if err != nil {
				return fmt.Errorf("failed to create incident for client %s: %w", client.ID[:8], err)
			}
			totalIncidents++
		}
	}

	fmt.Printf("✅ Successfully seeded %d incidents\n", totalIncidents)
	return nil
}

func createIncident(ctx context.Context, store *db.Store, client ClientInfo) error {
	// Generate incident date (within last 90 days)
	daysAgo := rand.Intn(90)
	incidentDate := pgtype.Date{
		Time:  time.Now().AddDate(0, 0, -daysAgo),
		Valid: true,
	}

	// Generate incident time (between 6:00 and 23:00)
	hour := 6 + rand.Intn(17)
	minute := rand.Intn(60)
	incidentTime := pgtype.Time{
		Microseconds: int64(hour*3600+minute*60) * 1000000,
		Valid:        true,
	}

	// Random incident type
	incidentType := randomElement(incidentTypes)

	// Get description based on type
	descriptions := incidentDescriptions[incidentType]
	description := randomElement(descriptions)

	// Get action taken based on type
	actions := actionsTaken[incidentType]
	actionTaken := randomElement(actions)

	// Random severity and status
	severity := randomElement(incidentSeverities)
	status := randomElement(incidentStatuses)

	// Other parties (optional)
	var otherPartiesStr *string
	party := randomElement(otherParties)
	if party != "" {
		otherPartiesStr = &party
	}

	// Generate ID
	incidentID, err := gonanoid.New()
	if err != nil {
		return fmt.Errorf("failed to generate incident ID: %w", err)
	}

	err = store.CreateIncident(ctx, db.CreateIncidentParams{
		ID:                  incidentID,
		ClientID:            client.ID,
		IncidentDate:        incidentDate,
		IncidentTime:        incidentTime,
		IncidentType:        incidentType,
		IncidentSeverity:    severity,
		LocationID:          client.LocationID,
		CoordinatorID:       client.CoordinatorID,
		IncidentDescription: description,
		ActionTaken:         actionTaken,
		OtherParties:        otherPartiesStr,
		Status:              status,
	})

	return err
}

// ============================================================
// Evaluations Seeding
// ============================================================

var (
	// Goal progress statuses
	goalStatuses = []db.GoalProgressStatus{
		db.GoalProgressStatusStarting,
		db.GoalProgressStatusOnTrack,
		db.GoalProgressStatusOnTrack, // More likely
		db.GoalProgressStatusDelayed,
		db.GoalProgressStatusAchieved,
	}

	// Progress notes corresponding to statuses (Dutch)
	progressNotes = map[db.GoalProgressStatus][]string{
		db.GoalProgressStatusStarting: {
			"Cliënt is enthousiast om aan dit doel te werken",
			"Eerste stappen zijn gezet, goede motivatie aanwezig",
			"Doel is besproken, concrete acties afgesproken",
			"Cliënt toont interesse maar heeft nog begeleiding nodig",
		},
		db.GoalProgressStatusOnTrack: {
			"Goede voortgang, cliënt houdt zich aan afspraken",
			"Positieve ontwikkeling zichtbaar, plan wordt gevolgd",
			"Doel verloopt volgens planning",
			"Cliënt laat groei zien op dit gebied",
			"Stappen worden genomen in de goede richting",
		},
		db.GoalProgressStatusDelayed: {
			"Voortgang loopt vertraging op door externe omstandigheden",
			"Cliënt heeft moeite met motivatie, extra ondersteuning geboden",
			"Doel blijkt complexer dan verwacht, plan wordt aangepast",
			"Terugval ervaren, herstelplan opgesteld",
		},
		db.GoalProgressStatusAchieved: {
			"Doel behaald! Cliënt kan dit nu zelfstandig",
			"Uitstekend resultaat, doel volledig gerealiseerd",
			"Mijlpaal bereikt, cliënt erg trots",
			"Met succes afgerond, kan focus op volgend doel",
		},
	}

	// Overall evaluation notes (Dutch)
	overallNotes = []string{
		"Cliënt heeft over het algemeen goede voortgang gemaakt in deze periode.",
		"Positieve evaluatie, cliënt toont groei en zelfinzicht.",
		"Periode met ups en downs, maar alles bij elkaar vooruitgang zichtbaar.",
		"Cliënt werkt goed mee aan behandelplan, goede samenwerking.",
		"Focus ligt nu op stabilisatie en behoud van behaalde resultaten.",
		"Enkele uitdagingen ervaren maar cliënt blijft gemotiveerd.",
		"Stabiele periode, duidelijke verbeteringen in dagstructuur.",
		"Goede balans gevonden tussen zelfstandigheid en ondersteuning.",
		"",
	}
)

func seedEvaluations(ctx context.Context, store *db.Store, inCareClients []ClientInfo) error {
	fmt.Println("🌱 Seeding evaluations for in_care clients...")

	totalEvaluations := 0

	for _, client := range inCareClients {
		// Get client details to retrieve goals and evaluation interval
		clientDetails, err := store.GetClientByID(ctx, client.ID)
		if err != nil {
			return fmt.Errorf("failed to get client details for %s: %w", client.ID, err)
		}

		// Get client goals
		goals, err := store.ListGoalsByClientID(ctx, &client.ID)
		if err != nil {
			return fmt.Errorf("failed to get goals for client %s: %w", client.ID, err)
		}

		if len(goals) == 0 {
			fmt.Printf("  ⚠ Client %s has no goals, skipping evaluations\n", client.ID)
			continue
		}

		// Determine evaluation interval (default to 5 weeks if not set)
		interval := int32(5)
		if clientDetails.EvaluationIntervalWeeks != nil {
			interval = *clientDetails.EvaluationIntervalWeeks
		}

		// Create 1-3 historical evaluations for each client
		numEvaluations := 1 + rand.Intn(3) // 1 to 3 evaluations

		for i := 0; i < numEvaluations; i++ {
			// Calculate evaluation date going backwards from today
			// Most recent evaluation is 0-2 weeks ago, then interval weeks back for each
			weeksAgo := i * int(interval)
			if i == 0 {
				// Most recent: 0-14 days ago
				weeksAgo = rand.Intn(15)
			} else {
				// Add some randomness (+/- 1 week)
				variation := rand.Intn(15) - 7
				weeksAgo += variation
			}

			evaluationDate := time.Now().AddDate(0, 0, -weeksAgo)

			// Create evaluation with goal progress logs
			if err := createEvaluationWithGoals(
				ctx,
				store,
				client.ID,
				client.CoordinatorID,
				evaluationDate,
				goals,
				interval,
				i == 0, // isLatest - update next_evaluation_date for most recent only
			); err != nil {
				return fmt.Errorf("failed to create evaluation for client %s: %w", client.ID, err)
			}

			totalEvaluations++
		}

		fmt.Printf("  ✓ Created %d evaluation(s) for client %s\n", numEvaluations, client.ID)
	}

	fmt.Printf("✅ Successfully seeded %d evaluations\n", totalEvaluations)
	return nil
}

func createEvaluationWithGoals(
	ctx context.Context,
	store *db.Store,
	clientID, coordinatorID string,
	evaluationDate time.Time,
	goals []db.ClientGoal,
	interval int32,
	isLatest bool,
) error {
	// Generate evaluation ID
	evaluationID, err := gonanoid.New()
	if err != nil {
		return fmt.Errorf("failed to generate evaluation ID: %w", err)
	}

	// Create progress logs for each goal
	progressLogs := make([]db.CreateGoalProgressLogParams, len(goals))
	for i, goal := range goals {
		progressLogID, err := gonanoid.New()
		if err != nil {
			return fmt.Errorf("failed to generate progress log ID: %w", err)
		}

		// Random status for this goal
		status := randomElement(goalStatuses)

		// Get a matching progress note for the status
		var progressNote *string
		if notes, ok := progressNotes[status]; ok && len(notes) > 0 {
			note := randomElement(notes)
			progressNote = &note
		}

		progressLogs[i] = db.CreateGoalProgressLogParams{
			ID:            progressLogID,
			EvaluationID:  evaluationID, // Will be set in transaction
			GoalID:        goal.ID,
			Status:        status,
			ProgressNotes: progressNote,
		}
	}

	// Random overall notes
	var overallNotesStr *string
	note := randomElement(overallNotes)
	if note != "" {
		overallNotesStr = &note
	}

	// Call the evaluation transaction (same as the real flow)
	_, err = store.CreateEvaluationTx(ctx, db.CreateEvaluationTxParams{
		Evaluation: db.CreateClientEvaluationParams{
			ID:             evaluationID,
			ClientID:       clientID,
			CoordinatorID:  coordinatorID,
			EvaluationDate: pgtype.Date{Time: evaluationDate, Valid: true},
			OverallNotes:   overallNotesStr,
			Status:         db.EvaluationStatusEnumSubmitted,
		},
		ProgressLogs:  progressLogs,
		IntervalWeeks: interval,
	})

	return err
}

// ============================================================
// Appointments Seeding
// ============================================================

// Appointment sample data
var (
	appointmentTitles = []string{
		"Intake gesprek",
		"Voortgangsgesprek",
		"Evaluatie bespreking",
		"Medicatie controle",
		"Familie bijeenkomst",
		"Teamoverleg cliënt",
		"Zorgplan bespreking",
		"Huisbezoek",
		"Ambulante begeleiding",
		"Crisisinterventie overleg",
	}

	appointmentDescriptions = []string{
		"Reguliere afspraak met cliënt voor voortgangsbespreking",
		"Bespreking van huidige situatie en doelen",
		"Evaluatie van behaalde resultaten afgelopen periode",
		"Controle medicatie-inname en bijwerkingen",
		"Gesprek met cliënt en familieleden over zorgplan",
		"Teamoverleg over voortgang cliënt",
		"Aanpassing zorgplan na evaluatie",
		"Bezoek aan cliënt thuis ter ondersteuning",
		"Ambulante begeleiding sessie",
		"Overleg naar aanleiding van crisissituatie",
	}

	appointmentLocations = []string{
		"Kantoor begeleider",
		"Woonkamer locatie",
		"Cliënt thuis",
		"Spreekkamer 1",
		"Spreekkamer 2",
		"Online via videobellen",
		"Gemeenschappelijke ruimte",
	}
)

func seedAppointments(
	ctx context.Context,
	store *db.Store,
	adminEmail string,
	inCareClients []ClientInfo,
) error {
	if adminEmail == "" {
		fmt.Println("⚠️  ADMIN_EMAIL not set, skipping appointment seeding")
		return nil
	}

	fmt.Printf("🌱 Seeding appointments for admin user: %s...\n", adminEmail)

	// Get admin user by email
	user, err := store.GetUserByEmail(ctx, adminEmail)
	if err != nil {
		return fmt.Errorf("failed to get admin user by email %s: %w", adminEmail, err)
	}

	// Get employee by user ID
	employee, err := store.GetEmployeeByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get employee for user %s: %w", user.ID, err)
	}

	organizerID := employee.ID
	now := time.Now()

	// Create appointments for the next 30 days
	appointmentCount := 0
	for i := 0; i < 15; i++ {
		appointmentID, err := gonanoid.New()
		if err != nil {
			return fmt.Errorf("failed to generate appointment ID: %w", err)
		}

		// Random offset in days (0-30) and hours (8-18)
		daysOffset := rand.Intn(30)
		hourOffset := 8 + rand.Intn(10) // Between 8:00 and 18:00
		startTime := now.AddDate(0, 0, daysOffset).Truncate(24 * time.Hour).Add(time.Duration(hourOffset) * time.Hour)
		endTime := startTime.Add(time.Duration(30+rand.Intn(60)) * time.Minute) // 30-90 min duration

		title := randomElement(appointmentTitles)
		description := randomElement(appointmentDescriptions)
		location := randomElement(appointmentLocations)

		// Random appointment type
		appointmentTypes := []db.AppointmentTypeEnum{
			db.AppointmentTypeEnumGeneral,
			db.AppointmentTypeEnumIntake,
			db.AppointmentTypeEnumAmbulatory,
		}
		appointmentType := randomElement(appointmentTypes)

		// Create appointment
		_, err = store.CreateAppointment(ctx, db.CreateAppointmentParams{
			ID:          appointmentID,
			Title:       title,
			Description: &description,
			StartTime:   pgtype.Timestamptz{Time: startTime, Valid: true},
			EndTime:     pgtype.Timestamptz{Time: endTime, Valid: true},
			Location:    &location,
			OrganizerID: organizerID,
			Status: db.NullAppointmentStatusEnum{
				AppointmentStatusEnum: db.AppointmentStatusEnumConfirmed,
				Valid:                 true,
			},
			Type: appointmentType,
		})
		if err != nil {
			return fmt.Errorf("failed to create appointment: %w", err)
		}

		// Add a client participant if available
		if len(inCareClients) > 0 {
			client := inCareClients[rand.Intn(len(inCareClients))]
			err = store.AddAppointmentParticipant(ctx, db.AddAppointmentParticipantParams{
				AppointmentID:   appointmentID,
				ParticipantID:   client.ID,
				ParticipantType: db.ParticipantTypeEnumClient,
			})
			if err != nil {
				return fmt.Errorf("failed to add appointment participant: %w", err)
			}
		}

		appointmentCount++
		fmt.Printf("  ✓ Created appointment: %s (%s)\n", title, startTime.Format("2006-01-02 15:04"))
	}

	fmt.Printf("✅ Successfully seeded %d appointments for admin\n", appointmentCount)
	return nil
}

// ============================================================
// Notifications Seeding
// ============================================================

// Notification templates for realistic scenarios
var notificationTemplates = []struct {
	Type           db.NotificationTypeEnum
	Priority       db.NotificationPriorityEnum
	TitlePattern   string
	MessagePattern string
}{
	// Incident notifications
	{
		Type:           db.NotificationTypeEnumIncidentCreated,
		Priority:       db.NotificationPriorityEnumHigh,
		TitlePattern:   "Nieuw incident gemeld",
		MessagePattern: "Er is een nieuw incident gemeld bij %s. Ernst: %s. Actie vereist.",
	},
	{
		Type:           db.NotificationTypeEnumIncidentCreated,
		Priority:       db.NotificationPriorityEnumUrgent,
		TitlePattern:   "Urgent: Ernstig incident",
		MessagePattern: "Er is een ernstig incident gemeld bij %s. Directe aandacht vereist.",
	},
	// Evaluation notifications
	{
		Type:           db.NotificationTypeEnumEvaluationDue,
		Priority:       db.NotificationPriorityEnumNormal,
		TitlePattern:   "Evaluatie binnenkort",
		MessagePattern: "De evaluatie voor cliënt %s staat gepland voor over %d dagen.",
	},
	{
		Type:           db.NotificationTypeEnumEvaluationDue,
		Priority:       db.NotificationPriorityEnumHigh,
		TitlePattern:   "Evaluatie morgen!",
		MessagePattern: "Herinnering: De evaluatie voor cliënt %s is morgen gepland.",
	},
	// Location transfer notifications
	{
		Type:           db.NotificationTypeEnumLocationTransferApproved,
		Priority:       db.NotificationPriorityEnumNormal,
		TitlePattern:   "Overplaatsing goedgekeurd",
		MessagePattern: "De overplaatsingsaanvraag voor cliënt %s is goedgekeurd.",
	},
	{
		Type:           db.NotificationTypeEnumLocationTransferRejected,
		Priority:       db.NotificationPriorityEnumNormal,
		TitlePattern:   "Overplaatsing afgewezen",
		MessagePattern: "De overplaatsingsaanvraag voor cliënt %s is afgewezen. Reden: %s",
	},
	// Appointment notifications
	{
		Type:           db.NotificationTypeEnumAppointmentReminder,
		Priority:       db.NotificationPriorityEnumNormal,
		TitlePattern:   "Afspraak vandaag",
		MessagePattern: "U heeft vandaag om %s een afspraak: %s",
	},
	{
		Type:           db.NotificationTypeEnumAppointmentReminder,
		Priority:       db.NotificationPriorityEnumLow,
		TitlePattern:   "Afspraak morgen",
		MessagePattern: "Herinnering: Morgen om %s heeft u een afspraak: %s",
	},
	// Client status notifications
	{
		Type:           db.NotificationTypeEnumClientStatusChange,
		Priority:       db.NotificationPriorityEnumNormal,
		TitlePattern:   "Cliëntstatus gewijzigd",
		MessagePattern: "De status van cliënt %s is gewijzigd naar '%s'.",
	},
	// System alerts
	{
		Type:           db.NotificationTypeEnumSystemAlert,
		Priority:       db.NotificationPriorityEnumHigh,
		TitlePattern:   "Systeemmelding",
		MessagePattern: "Gepland onderhoud: Het systeem zal op %s tijdelijk niet beschikbaar zijn.",
	},
}

// Client names for realistic notifications
var notificationClientNames = []string{
	"Jan de Vries", "Sophie Jansen", "Piet Bakker", "Emma Visser",
	"Lucas Smit", "Daan Meijer", "Lotte de Groot", "Finn Mulder",
}

// Location names for notifications
var notificationLocationNames = []string{
	"Woonlocatie De Zonnetuin", "Huize De Linde", "Villa Sereniteit",
	"Woongroep Horizon", "Locatie Oost", "Wooncentrum De Haven",
}

// Appointment titles for notifications
var notificationAppointmentTitles = []string{
	"Teamoverleg", "Evaluatiegesprek", "Intake nieuwe cliënt",
	"Voortgangsgesprek", "Behandelplanbespreking", "Multidisciplinair overleg",
}

func seedNotifications(ctx context.Context, store *db.Store) error {
	fmt.Println("🌱 Seeding notifications for admin users...")

	// Get all admin users
	adminUserIDs, err := store.GetUserIDsByRoleName(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to get admin users: %w", err)
	}

	if len(adminUserIDs) == 0 {
		fmt.Println("  ⚠ No admin users found, skipping notification seeding")
		return nil
	}

	fmt.Printf("  Found %d admin users\n", len(adminUserIDs))

	// Create notifications for each admin user
	notificationCount := 0
	for _, userID := range adminUserIDs {
		// Create 10-15 notifications per admin
		numNotifications := 10 + rand.Intn(6)

		for i := 0; i < numNotifications; i++ {
			if err := createRandomNotification(ctx, store, userID); err != nil {
				return fmt.Errorf("failed to create notification: %w", err)
			}
			notificationCount++
		}
	}

	fmt.Printf("✅ Successfully seeded %d notifications for admin users\n", notificationCount)
	return nil
}

func createRandomNotification(ctx context.Context, store *db.Store, userID string) error {
	// Generate notification ID
	notificationID, err := gonanoid.New()
	if err != nil {
		return fmt.Errorf("failed to generate notification ID: %w", err)
	}

	// Pick a random template
	template := randomElement(notificationTemplates)

	// Generate title and message based on template type
	var title, message string
	var resourceType, resourceID *string

	switch template.Type {
	case db.NotificationTypeEnumIncidentCreated:
		location := randomElement(notificationLocationNames)
		severity := randomElement([]string{"Minor", "Moderate", "Severe"})
		title = template.TitlePattern
		message = fmt.Sprintf(template.MessagePattern, location, severity)
		rt, rid := "incident", generateFakeID()
		resourceType, resourceID = &rt, &rid

	case db.NotificationTypeEnumEvaluationDue:
		client := randomElement(notificationClientNames)
		days := rand.Intn(7) + 1
		title = template.TitlePattern
		if days == 1 {
			message = fmt.Sprintf("Herinnering: De evaluatie voor cliënt %s is morgen gepland.", client)
		} else {
			message = fmt.Sprintf(template.MessagePattern, client, days)
		}
		rt, rid := "client", generateFakeID()
		resourceType, resourceID = &rt, &rid

	case db.NotificationTypeEnumLocationTransferApproved:
		client := randomElement(notificationClientNames)
		title = template.TitlePattern
		message = fmt.Sprintf(template.MessagePattern, client)
		rt, rid := "location_transfer", generateFakeID()
		resourceType, resourceID = &rt, &rid

	case db.NotificationTypeEnumLocationTransferRejected:
		client := randomElement(notificationClientNames)
		reasons := []string{"Geen beschikbare plek", "Zorgvraag niet passend", "Wachtlijst vol"}
		title = template.TitlePattern
		message = fmt.Sprintf(template.MessagePattern, client, randomElement(reasons))
		rt, rid := "location_transfer", generateFakeID()
		resourceType, resourceID = &rt, &rid

	case db.NotificationTypeEnumAppointmentReminder:
		appointmentTitle := randomElement(notificationAppointmentTitles)
		appointmentTime := fmt.Sprintf("%02d:%02d", 8+rand.Intn(10), rand.Intn(4)*15)
		title = template.TitlePattern
		message = fmt.Sprintf(template.MessagePattern, appointmentTime, appointmentTitle)
		rt, rid := "appointment", generateFakeID()
		resourceType, resourceID = &rt, &rid

	case db.NotificationTypeEnumClientStatusChange:
		client := randomElement(notificationClientNames)
		statuses := []string{"In zorg", "Wachtlijst", "Uitstroom gepland"}
		title = template.TitlePattern
		message = fmt.Sprintf(template.MessagePattern, client, randomElement(statuses))
		rt, rid := "client", generateFakeID()
		resourceType, resourceID = &rt, &rid

	case db.NotificationTypeEnumSystemAlert:
		dates := []string{"maandag 9:00-11:00", "woensdag 22:00-23:00", "zaterdag 05:00-07:00"}
		title = template.TitlePattern
		message = fmt.Sprintf(template.MessagePattern, randomElement(dates))
		// No resource for system alerts

	default:
		title = "Melding"
		message = "Er is een nieuwe melding voor u."
	}

	// Create the notification
	_, err = store.CreateNotification(ctx, db.CreateNotificationParams{
		ID:           notificationID,
		UserID:       userID,
		Type:         template.Type,
		Priority:     template.Priority,
		Title:        title,
		Message:      message,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func generateFakeID() string {
	id, _ := gonanoid.New()
	return id
}

// ============================================================
// Audit Logs Seeding
// ============================================================

// Realistic resource types that match API endpoints
var auditResourceTypes = []string{
	"client", "employee", "incident", "evaluation", "appointment",
	"location", "location_transfer", "notification", "referring_org",
	"registration_form", "intake_form",
}

// Realistic user agents for different devices
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPad; CPU OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 14; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.43 Mobile Safari/537.36",
}

// Realistic failure reasons
var failureReasons = []string{
	"unauthorized", "forbidden", "not_found", "validation_error",
	"server_error", "rate_limited", "connection_timeout",
}

// Action distribution - reads are most common
var actionWeights = map[db.AuditActionEnum]int{
	db.AuditActionEnumRead:   60, // 60% reads
	db.AuditActionEnumCreate: 15, // 15% creates
	db.AuditActionEnumUpdate: 15, // 15% updates
	db.AuditActionEnumDelete: 5,  // 5% deletes
	db.AuditActionEnumLogin:  3,  // 3% logins
	db.AuditActionEnumLogout: 2,  // 2% logouts
}

func seedAuditLogs(
	ctx context.Context,
	store *db.Store,
	userIDs, employeeIDs, clientIDs []string,
	daysBack int,
	count int,
) error {
	fmt.Printf("🌱 Seeding %d audit logs over the last %d days...\n", count, daysBack)

	if len(userIDs) == 0 {
		fmt.Println("  ⚠ No users found, skipping audit log seeding")
		return nil
	}

	prevHash := "GENESIS"
	var lastCreatedAt time.Time

	for i := 0; i < count; i++ {
		// Select random user and employee
		userID := randomElement(userIDs)
		var employeeID *string
		if len(employeeIDs) > 0 && rand.Float32() < 0.9 {
			empID := randomElement(employeeIDs)
			employeeID = &empID
		}

		// Select random action based on weights
		action := selectWeightedAction()

		// Select random resource type
		resourceType := randomElement(auditResourceTypes)

		// Generate timestamp spread over the past N days
		daysAgo := rand.Intn(daysBack)
		hoursAgo := rand.Intn(24)
		minutesAgo := rand.Intn(60)
		createdAt := time.Now().AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursAgo)*time.Hour - time.Duration(minutesAgo)*time.Minute)

		// Ensure timestamps are sequential
		if !lastCreatedAt.IsZero() && createdAt.After(lastCreatedAt) {
			createdAt = lastCreatedAt.Add(time.Second)
		}
		lastCreatedAt = createdAt

		// Most operations succeed, small chance of failure
		isFailure := rand.Float32() < 0.05
		var status db.AuditStatusEnum
		var failureReason *string
		if isFailure {
			status = db.AuditStatusEnumFailure
			fr := randomElement(failureReasons)
			failureReason = &fr
		} else {
			status = db.AuditStatusEnumSuccess
		}

		// Generate resource ID for non-read operations
		var resourceID *string
		if action != db.AuditActionEnumRead && action != db.AuditActionEnumLogin && action != db.AuditActionEnumLogout {
			rid := generateFakeID()
			resourceID = &rid
		}

		// Generate client ID for client-related operations
		var clientID *string
		if resourceType == "client" && len(clientIDs) > 0 && rand.Float32() < 0.8 {
			cid := randomElement(clientIDs)
			clientID = &cid
		}

		// Generate IP address (realistic Dutch IP ranges)
		ipAddress := fmt.Sprintf("10.0.%d.%d", rand.Intn(256), rand.Intn(256))
		if rand.Float32() < 0.3 {
			ipAddress = fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256))
		}
		if rand.Float32() < 0.1 {
			ipAddress = fmt.Sprintf("213.34.%d.%d", rand.Intn(256), rand.Intn(256))
		}

		// Generate request ID
		requestID := generateFakeID()

		// Generate old/new values for updates/deletes
		var oldValue, newValue []byte
		if action == db.AuditActionEnumUpdate || action == db.AuditActionEnumDelete {
			oldData := map[string]interface{}{
				"updated_at": createdAt.Add(-24 * time.Hour).Format(time.RFC3339),
				"version":    rand.Intn(5) + 1,
			}
			oldValue, _ = json.Marshal(oldData)
		}
		if action == db.AuditActionEnumCreate || action == db.AuditActionEnumUpdate {
			newData := map[string]interface{}{
				"created_at": createdAt.Format(time.RFC3339),
				"created_by": userID,
				"version":    rand.Intn(5) + 1,
			}
			newValue, _ = json.Marshal(newData)
		}

		// Generate audit log entry
		id := generateFakeID()
		currentHash := computeAuditHash(
			id,
			strPtr(userID),
			employeeID,
			string(action),
			resourceType,
			resourceID,
			oldValue,
			newValue,
			&ipAddress,
			strPtr(requestID),
			status,
			failureReason,
			prevHash,
			createdAt,
		)

		// Create the audit log entry
		err := store.CreateAuditLog(ctx, db.CreateAuditLogParams{
			ID:            id,
			UserID:        strPtr(userID),
			EmployeeID:    employeeID,
			ClientID:      clientID,
			Action:        action,
			ResourceType:  resourceType,
			ResourceID:    resourceID,
			OldValue:      oldValue,
			NewValue:      newValue,
			IpAddress:     strPtr(ipAddress),
			UserAgent:     strPtr(randomElement(userAgents)),
			RequestID:     strPtr(requestID),
			Status:        status,
			FailureReason: failureReason,
			PrevHash:      prevHash,
			CurrentHash:   currentHash,
		})
		if err != nil {
			return fmt.Errorf("failed to create audit log %d: %w", i+1, err)
		}

		prevHash = currentHash
	}

	fmt.Printf("✅ Successfully seeded %d audit logs\n", count)
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func selectWeightedAction() db.AuditActionEnum {
	total := 0
	for _, weight := range actionWeights {
		total += weight
	}

	randVal := rand.Intn(total)
	cumulative := 0

	for action, weight := range actionWeights {
		cumulative += weight
		if randVal < cumulative {
			return action
		}
	}

	return db.AuditActionEnumRead
}

// computeAuditHash generates a SHA-256 hash matching the middleware's logic
func computeAuditHash(
	id string,
	userID, employeeID *string,
	action, resourceType string,
	resourceID *string,
	oldValue, newValue []byte,
	ipAddress, requestID *string,
	status db.AuditStatusEnum,
	failureReason *string,
	prevHash string,
	createdAt time.Time,
) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		id,
		strVal(userID),
		strVal(employeeID),
		action,
		resourceType,
		strVal(resourceID),
		string(oldValue),
		string(newValue),
		strVal(ipAddress),
		strVal(requestID),
		string(status),
		strVal(failureReason),
		prevHash,
		createdAt.UTC().Format(time.RFC3339Nano),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func strVal(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
