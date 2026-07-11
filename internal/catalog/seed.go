package catalog

// This is the seeded bike-shop catalog — a small, believable dataset (à la a local independent shop). Product
// images are rendered by the storefront as deterministic placeholders keyed on the product ID, so there are no
// binary assets to ship. Prices are in USD cents.

func seedCategories() []Category {
	return []Category{
		{Slug: "road", Name: "Road Bikes", Kind: KindBikes, Blurb: "Fast on the tarmac — endurance to race."},
		{Slug: "gravel", Name: "Gravel Bikes", Kind: KindBikes, Blurb: "One bike, every surface."},
		{Slug: "mountain", Name: "Mountain Bikes", Kind: KindBikes, Blurb: "Trail, all-mountain, and hardtail."},
		{Slug: "commuter", Name: "Commuter & Urban", Kind: KindBikes, Blurb: "Get to work, rain or shine."},
		{Slug: "kids", Name: "Kids' Bikes", Kind: KindBikes, Blurb: "First pedals to first trails."},
		{Slug: "ebike-city", Name: "Electric City", Kind: KindEbikes, Blurb: "Pedal-assist for the daily ride."},
		{Slug: "ebike-mountain", Name: "Electric Mountain", Kind: KindEbikes, Blurb: "More trail, less shuttle."},
		{Slug: "drivetrain", Name: "Drivetrain", Kind: KindComponents, Blurb: "Shift and stop with confidence."},
		{Slug: "wheelsets", Name: "Wheelsets", Kind: KindWheels, Blurb: "Lighter, stiffer, faster."},
		{Slug: "tires", Name: "Tires & Tubes", Kind: KindTires, Blurb: "Grip for every condition."},
		{Slug: "accessories", Name: "Accessories", Kind: KindAccessories, Blurb: "Lights, locks, bags, and tools."},
		{Slug: "apparel", Name: "Apparel & Helmets", Kind: KindApparel, Blurb: "Ride comfortable, ride safe."},
	}
}

func seedBrands() []Brand {
	return []Brand{
		{Slug: "all-city", Name: "All-City"},
		{Slug: "cannondale", Name: "Cannondale"},
		{Slug: "kona", Name: "Kona"},
		{Slug: "marin", Name: "Marin"},
		{Slug: "salsa", Name: "Salsa"},
		{Slug: "specialized", Name: "Specialized"},
		{Slug: "surly", Name: "Surly"},
	}
}

func seedProducts() []Product {
	return []Product{
		// --- Road ---
		{ID: "p-1001", Slug: "cannondale-synapse-carbon-2", Name: "Synapse Carbon 2 RLE", Brand: "cannondale", Category: "road",
			PriceCents: 420000, Summary: "Endurance carbon road bike with SRAM Rival eTap AXS.", Featured: true,
			Description: "The Synapse is Cannondale's endurance road platform — a light carbon frame with SmartSense integrated lighting and radar, room for 35mm tires, and all-day comfort geometry.",
			Sizes:      []string{"51", "54", "56", "58"}, InStock: true,
			Specs: map[string]string{"Frame": "BallisTec Carbon", "Groupset": "SRAM Rival eTap AXS", "Wheels": "RD 2.0", "Weight": "8.9 kg"}},
		{ID: "p-1002", Slug: "specialized-allez-sprint", Name: "Allez Sprint Comp", Brand: "specialized", Category: "road",
			PriceCents: 320000, SalePriceCents: 279900, Summary: "Aero alloy crit weapon with Shimano 105 Di2.",
			Description: "The Allez Sprint brings Tarmac-derived aero tube shapes to a smart-welded alloy frame — a race bike that shrugs off curbs and crashes.",
			Sizes:      []string{"49", "52", "54", "56"}, InStock: true,
			Specs: map[string]string{"Frame": "Smartweld E5 Alloy", "Groupset": "Shimano 105 Di2", "Wheels": "DT R470"}},
		{ID: "p-1003", Slug: "all-city-zig-zag", Name: "Zig Zag", Brand: "all-city", Category: "road",
			PriceCents: 289900, Summary: "Steel all-road bike with a lively ride and classic looks.",
			Description: "A-C's stainless steel road bike: sharp handling, clearance for 35mm rubber, and the timeless ride quality only good steel delivers.",
			Sizes:      []string{"52", "55", "58"}, InStock: true,
			Specs: map[string]string{"Frame": "A-C 612 Select Stainless", "Groupset": "Shimano 105", "Tire clearance": "35mm"}},

		// --- Gravel ---
		{ID: "p-1010", Slug: "salsa-warbird-c-grx", Name: "Warbird C GRX 810", Brand: "salsa", Category: "gravel",
			PriceCents: 399900, Summary: "Carbon gravel race bike built for the long haul.", Featured: true,
			Description: "The Warbird is Salsa's dedicated gravel racer — Class 5 VRS compliance to smooth the chatter, room for 45mm tires, and multiple bottle mounts for all-day epics.",
			Sizes:      []string{"52.5", "55", "57.5"}, InStock: true,
			Specs: map[string]string{"Frame": "Carbon w/ Class 5 VRS", "Groupset": "Shimano GRX 810", "Tire clearance": "45mm 700c"}},
		{ID: "p-1011", Slug: "surly-midnight-special", Name: "Midnight Special", Brand: "surly", Category: "gravel",
			PriceCents: 254900, Summary: "Steel all-road that loves pavement and dirt equally.", Featured: true,
			Description: "A road-plus bruiser: 650b x 60mm tires, low-trail-ish handling, and Surly's bombproof steel. The one-bike quiver-killer.",
			Sizes:      []string{"46", "52", "56", "60"}, InStock: true,
			Specs: map[string]string{"Frame": "Surly 4130 CroMoly", "Groupset": "Shimano GRX 400", "Tire clearance": "60mm 650b"}},
		{ID: "p-1012", Slug: "kona-libre", Name: "Libre", Brand: "kona", Category: "gravel",
			PriceCents: 329900, SalePriceCents: 289900, Summary: "Carbon adventure gravel with bikepacking mounts everywhere.",
			Description: "Kona's do-it-all gravel bike: relaxed geometry, mounts for the whole kitchen, and clearance for 700x45c or 650b x 2.2in.",
			Sizes:      []string{"52", "54", "56"}, InStock: true,
			Specs: map[string]string{"Frame": "Kona Carbon", "Groupset": "Shimano GRX", "Mounts": "Fork + top tube + triple boss"}},

		// --- Mountain ---
		{ID: "p-1020", Slug: "kona-process-134", Name: "Process 134 CR/DL 29", Brand: "kona", Category: "mountain",
			PriceCents: 549900, Summary: "Carbon trail bike that punches well above its travel.", Featured: true,
			Description: "134mm rear / 140mm front of pure trail fun. The Process descends like a bike with more travel and climbs like one with less.",
			Sizes:      []string{"S", "M", "L", "XL"}, InStock: true,
			Specs: map[string]string{"Travel": "134/140mm", "Wheels": "29\"", "Groupset": "SRAM GX Eagle", "Fork": "RockShox Pike"}},
		{ID: "p-1021", Slug: "marin-rift-zone-2", Name: "Rift Zone 2 29", Brand: "marin", Category: "mountain",
			PriceCents: 229900, Summary: "Aluminum trail bike that overdelivers on value.",
			Description: "Marin's Rift Zone is the value benchmark for a modern 125mm trail bike — playful, capable, and priced to get you riding.",
			Sizes:      []string{"S", "M", "L", "XL"}, InStock: true,
			Specs: map[string]string{"Travel": "125/130mm", "Wheels": "29\"", "Groupset": "Shimano Deore 12s"}},
		{ID: "p-1022", Slug: "specialized-rockhopper-comp", Name: "Rockhopper Comp 29", Brand: "specialized", Category: "mountain",
			PriceCents: 109900, Summary: "The hardtail that has started a million trail habits.",
			Description: "A confident, modern hardtail with a dropper post and a tapered-headtube frame that's ready for real trails.",
			Sizes:      []string{"S", "M", "L"}, InStock: true,
			Specs: map[string]string{"Frame": "Premium Alloy", "Fork": "SR Suntour XCM", "Dropper": "Yes"}},

		// --- Commuter ---
		{ID: "p-1030", Slug: "marin-fairfax-1", Name: "Fairfax 1", Brand: "marin", Category: "commuter",
			PriceCents: 64900, Summary: "Fast, light hybrid for the daily commute.",
			Description: "Flat-bar comfort, rack and fender mounts, and quick-rolling 700c tires — the honest workhorse commuter.",
			Sizes:      []string{"S", "M", "L", "XL"}, InStock: true,
			Specs: map[string]string{"Frame": "Series 2 Aluminum", "Drivetrain": "Shimano 2x8", "Mounts": "Rack + fender"}},
		{ID: "p-1031", Slug: "surly-cross-check", Name: "Cross-Check", Brand: "surly", Category: "commuter",
			PriceCents: 174900, Summary: "The legendary do-anything steel workhorse.", Featured: true,
			Description: "Commute, tour, race cross, haul groceries — the Cross-Check has done it all for 20 years and isn't stopping.",
			Sizes:      []string{"46", "52", "56", "60"}, InStock: true,
			Specs: map[string]string{"Frame": "Surly 4130 CroMoly", "Drivetrain": "Shimano 2x9", "Tire clearance": "42mm"}},

		// --- Kids ---
		{ID: "p-1040", Slug: "specialized-jett-20", Name: "Jett 20", Brand: "specialized", Category: "kids",
			PriceCents: 46000, Summary: "Lightweight 20\" wheel bike that grows with the rider.",
			Description: "A genuinely light kids' bike with a low standover and a rider-fit-system saddle/bar that adjusts as they grow.",
			Sizes:      []string{"20\""}, InStock: true,
			Specs: map[string]string{"Wheels": "20\"", "Gears": "1x7", "Weight": "9.9 kg"}},

		// --- E-bikes ---
		{ID: "p-1050", Slug: "specialized-turbo-vado-40", Name: "Turbo Vado 4.0", Brand: "specialized", Category: "ebike-city",
			PriceCents: 385000, Summary: "Smooth, quiet pedal-assist commuter with 90+ mile range.", Featured: true,
			Description: "The Vado's custom Specialized 2.2 motor and 710Wh battery make hills disappear and errands effortless — integrated lights, rack, and fenders included.",
			Sizes:      []string{"S", "M", "L"}, InStock: true,
			Specs: map[string]string{"Motor": "Specialized 2.2 (90Nm)", "Battery": "710Wh", "Range": "Up to 90 mi", "Class": "3"}},
		{ID: "p-1051", Slug: "cannondale-moterra-neo", Name: "Moterra Neo 4", Brand: "cannondale", Category: "ebike-mountain",
			PriceCents: 545000, SalePriceCents: 489900, Summary: "Full-power e-MTB with Bosch Performance CX.",
			Description: "A do-it-all full-suspension e-MTB — 150mm travel, Bosch's punchy CX motor, and a 750Wh battery for lap after lap.",
			Sizes:      []string{"S", "M", "L", "XL"}, InStock: true,
			Specs: map[string]string{"Motor": "Bosch Performance CX (85Nm)", "Battery": "750Wh", "Travel": "150mm"}},

		// --- Drivetrain (components) ---
		{ID: "p-2001", Slug: "shimano-grx-rd-810", Name: "GRX RD-RX810 Rear Derailleur", Brand: "salsa", Category: "drivetrain",
			PriceCents: 11500, Summary: "11-speed gravel rear derailleur with clutch.",
			Description: "Shimano's GRX clutch derailleur keeps the chain quiet and secure over the rough stuff.",
			InStock: true, Specs: map[string]string{"Speeds": "11", "Clutch": "Shadow RD+"}},
		{ID: "p-2002", Slug: "sram-gx-eagle-cassette", Name: "GX Eagle XG-1275 Cassette", Brand: "specialized", Category: "drivetrain",
			PriceCents: 21000, SalePriceCents: 17900, Summary: "12-speed 10-52T wide-range MTB cassette.",
			Description: "A huge 10-52T range for steep climbs, at a weight and price that make sense.",
			InStock: true, Specs: map[string]string{"Speeds": "12", "Range": "10-52T"}},

		// --- Wheels ---
		{ID: "p-3001", Slug: "dt-swiss-gr1600", Name: "DT Swiss GR 1600 Gravel Wheelset", Brand: "salsa", Category: "wheelsets",
			PriceCents: 69900, Summary: "Tough, tubeless-ready 700c gravel wheels.", Featured: true,
			Description: "Reliable, serviceable, and tubeless-ready — 24mm internal width for modern gravel rubber.",
			InStock: true, Specs: map[string]string{"Size": "700c", "Internal width": "24mm", "Tubeless": "Yes"}},
		{ID: "p-3002", Slug: "roval-traverse-29", Name: "Traverse 29 Alloy Wheelset", Brand: "specialized", Category: "wheelsets",
			PriceCents: 55000, Summary: "Trail-tough 29\" alloy wheels, tubeless-ready.",
			Description: "Wide, stiff, and durable — the go-to alloy trail wheel upgrade.",
			InStock: false, Specs: map[string]string{"Size": "29\"", "Internal width": "29mm"}},

		// --- Tires ---
		{ID: "p-4001", Slug: "wtb-riddler-45", Name: "Riddler 45c TCS Gravel Tire", Brand: "all-city", Category: "tires",
			PriceCents: 6500, Summary: "Fast-rolling center with cornering knobs — a gravel favorite.",
			Description: "The Riddler's tightly-packed center tread hums on hardpack while the shoulder knobs bite in the corners.",
			InStock: true, Specs: map[string]string{"Size": "700x45c", "Tubeless": "TCS"}},
		{ID: "p-4002", Slug: "maxxis-minion-dhf-29", Name: "Minion DHF 29x2.5 Tire", Brand: "kona", Category: "tires",
			PriceCents: 8500, SalePriceCents: 6900, Summary: "The benchmark front trail/enduro tire.",
			Description: "Paddle-like cornering knobs and predictable braking — the DHF is the front tire everyone copies.",
			InStock: true, Specs: map[string]string{"Size": "29x2.5", "Compound": "3C MaxxTerra"}},

		// --- Accessories ---
		{ID: "p-5001", Slug: "kryptonite-evolution-lock", Name: "Evolution Mini-7 U-Lock", Brand: "surly", Category: "accessories",
			PriceCents: 8999, Summary: "Hardened U-lock with a 4-foot cable.",
			Description: "Sold Secure Gold hardened steel with a double-deadbolt — serious security for the daily park-up.",
			InStock: true, Specs: map[string]string{"Security": "Sold Secure Gold"}},
		{ID: "p-5002", Slug: "lezyne-macro-drive-light", Name: "Macro Drive 1400+ Front Light", Brand: "marin", Category: "accessories",
			PriceCents: 9500, Summary: "1400-lumen USB-C rechargeable headlight.", Featured: true,
			Description: "Bright enough to commute or night-ride, with a machined alloy body and USB-C charging.",
			InStock: true, Specs: map[string]string{"Output": "1400 lm", "Charge": "USB-C"}},
		{ID: "p-5003", Slug: "ortlieb-back-roller", Name: "Back-Roller Classic Panniers", Brand: "surly", Category: "accessories",
			PriceCents: 21000, Summary: "Legendary 100% waterproof commuter/touring panniers (pair).",
			Description: "The waterproof pannier standard — roll-top closure, QL2.1 mounts, and bombproof fabric.",
			InStock: true, Specs: map[string]string{"Volume": "40L pair", "Waterproof": "IP64"}},

		// --- Apparel ---
		{ID: "p-6001", Slug: "giro-synthe-mips-helmet", Name: "Synthe MIPS II Helmet", Brand: "specialized", Category: "apparel",
			PriceCents: 26000, Summary: "Light, aero, well-ventilated road helmet with MIPS.",
			Description: "Wind-tunnel-shaped ventilation and a rotational-impact MIPS liner — race protection you'll forget you're wearing.",
			Sizes:      []string{"S", "M", "L"}, InStock: true,
			Specs: map[string]string{"Protection": "MIPS", "Weight": "250g"}},
		{ID: "p-6002", Slug: "pearl-izumi-attack-jersey", Name: "Attack Air Jersey", Brand: "cannondale", Category: "apparel",
			PriceCents: 8500, SalePriceCents: 5900, Summary: "Breathable summer race-fit jersey.",
			Description: "Featherweight, quick-drying fabric and a race cut for hot-weather efforts.",
			Sizes:      []string{"S", "M", "L", "XL"}, InStock: true,
			Specs: map[string]string{"Fit": "Race", "UPF": "50+"}},
	}
}
