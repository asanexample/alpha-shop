import { Link } from "react-router-dom";
import { Breadcrumb } from "../components/Breadcrumb";
import styles from "./Info.module.css";

interface Section {
  heading: string;
  body: string;
}
interface InfoContent {
  eyebrow: string;
  title: string;
  lead: string;
  sections: Section[];
  cta?: { label: string; to: string };
}

const CONTENT: Record<string, InfoContent> = {
  service: {
    eyebrow: "The workshop",
    title: "Service & repair",
    lead: "Every bike we sell rolls out tuned. Every bike you bring in gets the same care.",
    sections: [
      {
        heading: "Book online or roll in",
        body: "Drop your bike at the bench any day we're open, or reserve a slot online. Standard tune-ups turn around next day; we'll call with a quote before we touch anything beyond the basics.",
      },
      {
        heading: "What we do",
        body: "Tune-ups, brake and drivetrain overhauls, hand-built wheels, suspension service, tubeless setups, and full frame-up custom builds. If it rolls, we'll work on it — road, gravel, mountain, or e-bike.",
      },
      {
        heading: "Custom builds",
        body: "Bring a frame or an idea. We'll spec the parts, source them, and build it to ride the way you want — then fit you to it on the jig before it leaves.",
      },
    ],
    cta: { label: "Shop bikes", to: "/c/road" },
  },
  about: {
    eyebrow: "Since 2009",
    title: "Our story",
    lead: "A repair bench that grew into a shop, still run by the people who wrench on your bike.",
    sections: [
      {
        heading: "Independent & rider-owned",
        body: "We opened on SE Belmont in 2009 with two work stands and a used cash register. Fifteen years later we're still independent, still rider-owned, and still more interested in getting you riding than moving inventory.",
      },
      {
        heading: "Bikes we actually ride",
        body: "We only stock brands we'd put our own miles on. That's why the range is tight — seven brands, chosen because they hold up in the wet Northwest, not because a distributor pushed them.",
      },
      {
        heading: "Visit us",
        body: "2340 SE Belmont St, Portland OR. Open Monday through Saturday, 10 to 7. Free parking out back, coffee on, dog-friendly.",
      },
    ],
    cta: { label: "Book a service", to: "/service" },
  },
  community: {
    eyebrow: "More than a store",
    title: "Community",
    lead: "The shop is a meeting point — for rides, for learning, and for keeping Portland rolling.",
    sections: [
      {
        heading: "Saturday shop ride",
        body: "A no-drop 25-mile spin leaves the front door every Saturday at 8am. All bikes, all paces welcome. Coffee and pastries after, always.",
      },
      {
        heading: "Wrench night",
        body: "First Tuesday of the month, we open the workstands for free hands-on repair classes. Bring your bike and learn to fix it yourself — flats, brakes, shifting, and more.",
      },
      {
        heading: "Giving back",
        body: "A share of every sale goes to local trail building and safe-streets advocacy. Riding is better when the places we ride are, too.",
      },
    ],
    cta: { label: "Join a ride", to: "/" },
  },
};

export function Info({ page }: { page: keyof typeof CONTENT }) {
  const c = CONTENT[page];
  if (!c) return null;
  return (
    <div className={styles.wrap}>
      <Breadcrumb items={[{ label: "Home", to: "/" }, { label: c.title }]} />
      <p className="eyebrow">{c.eyebrow}</p>
      <h1 className={styles.title}>{c.title}</h1>
      <p className={styles.lead}>{c.lead}</p>
      {c.sections.map((s) => (
        <section key={s.heading} className={styles.section}>
          <h2>{s.heading}</h2>
          <p>{s.body}</p>
        </section>
      ))}
      {c.cta ? (
        <p className={styles.cta}>
          <Link className="btn" to={c.cta.to}>
            {c.cta.label}
          </Link>
        </p>
      ) : null}
    </div>
  );
}
