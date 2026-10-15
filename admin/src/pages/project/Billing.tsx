import { request, type Data } from "../../api";
import { useI18n } from "../../context";
import { Button, Card, Loading, useLoad } from "../../ui";

export default function Billing() {
  const { tr } = useI18n();
  const [data] = useLoad<Data>(() => request("/billing"));
  if (!data) return <Loading />;
  return (
    <>
      <Card title={tr("billing.currentTitle")} description={tr("billing.currentDesc", { plan: data.plan_label, projects: data.max_projects, status: data.status })} />
      {data.public_beta && <div className="notice">🚀 <strong>{tr("billing.betaNotice")}</strong>{tr("billing.betaBody")}</div>}
      <div className="plans">
        {(data.plans || []).map((plan: Data) => (
          <Card key={plan.id} title={plan.label} className={plan.id === data.plan ? "current-plan" : ""}>
            <strong className="price">
              {data.public_beta
                ? (plan.id === "free" ? tr("billing.forever") : tr("billing.publicBeta"))
                : plan.price_try ? `${plan.price_try.amount_cents / 100} TRY` : tr("billing.free")}
            </strong>
            <ul>{(plan.features || []).map((feature: string) => <li key={feature}>{feature}</li>)}</ul>
            <Button primary={plan.id !== data.plan}>
              {plan.id === data.plan ? tr("billing.currentPlan") : data.public_beta ? tr("billing.includedBeta") : tr("billing.upgrade")}
            </Button>
          </Card>
        ))}
      </div>
    </>
  );
}
