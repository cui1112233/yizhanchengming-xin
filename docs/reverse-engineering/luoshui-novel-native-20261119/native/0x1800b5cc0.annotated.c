/* REA ev_f2f56a573b43e899f37d7b7def89a74cbd69a7be8f8600753795161c831287d2; original SHA b5b22086...; constant comments are analyst annotations. */

code * FUN_1800b5cc0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  longlong *plVar4;
  int *piVar5;
  longlong *plVar6;
  int iVar7;
  longlong *plVar8;
  longlong *plVar9;
  code *pcVar10;
  code *pcVar11;
  longlong lVar12;
  undefined8 uVar13;
  code *local_res20;
  
  plVar6 = DAT_1803b1f18;
  plVar1 = (longlong *)param_3[2];
  plVar2 = (longlong *)*param_3;
  plVar3 = (longlong *)param_3[1];
  plVar4 = (longlong *)param_3[3];
  local_res20 = (code *)0x0;
  if (DAT_1803b1f18 == (longlong *)0x0) {
LAB_1800b5d29:
    DAT_1803b1f18 = (longlong *)FUN_18036cd10(param_1,DAT_1803b1f40,DAT_1803ca7c0,0x20);
  }
  else if ((1 < *DAT_1803b1f18) || (DAT_1803b1f18[2] != 0)) {
    if ((DAT_1803b1f18 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b1f18 && (*DAT_1803b1f18 = *DAT_1803b1f18 + -1, *plVar6 == 0)))) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    goto LAB_1800b5d29;
  }
  plVar6 = DAT_1803b1f18;
  FUN_180001d00(param_1,DAT_1803b1f18);
  iVar7 = FUN_1803841f0(plVar3,DAT_1803ab250);
  pcVar11 = _Py_TrueStruct_exref;
  if (iVar7 == -1) {
    lVar12 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar13 = 0x2b;
  }
  else {
    if (iVar7 == 1) {
      if (*(int *)_Py_TrueStruct_exref + 1 != 0) {
        *(int *)_Py_TrueStruct_exref = *(int *)_Py_TrueStruct_exref + 1;
      }
      goto LAB_1800b6100;
    }
    iVar7 = (int)*plVar2 + 1;
    if (iVar7 != 0) {
      *(int *)plVar2 = iVar7;
    }
    plVar8 = (longlong *)FUN_180381a40(plVar3,plVar1);
    if (plVar8 == (longlong *)0x0) {
      lVar12 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      plVar9 = (longlong *)PyNumber_Long(plVar8);
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if (plVar9 != (longlong *)0x0) {
        pcVar10 = (code *)FUN_180383ba0(plVar9,plVar2);
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        if ((pcVar10 != (code *)0x0) &&
           (iVar7 = FUN_180001510(pcVar10), local_res20 = pcVar10, iVar7 != -1)) {
          if (iVar7 == 0) {
            pcVar11 = pcVar10;
            if (*(int *)pcVar10 + 1 != 0) {
              *(int *)pcVar10 = *(int *)pcVar10 + 1;
            }
LAB_1800b60d0:
            if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
              (**(code **)(plVar2[1] + 0x30))(plVar2);
            }
            if ((-1 < *(int *)pcVar10) &&
               (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
              (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
            }
LAB_1800b6100:
            lVar12 = **(longlong **)(param_1 + 0x38);
            plVar6 = *(longlong **)(lVar12 + 0x30);
            **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar12 + 8);
            *(undefined4 *)(plVar6 + 8) = 0xffffffff;
            plVar8 = (longlong *)plVar6[2];
            if (((plVar8 != (longlong *)0x0) && (plVar6[2] = 0, -1 < (int)*plVar8)) &&
               (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))();
            }
            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
              (**(code **)(plVar6[1] + 0x30))(plVar6);
            }
            plVar6[10] = 0;
            if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
              (**(code **)(plVar2[1] + 0x30))(plVar2);
            }
            if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
              (**(code **)(plVar3[1] + 0x30))(plVar3);
            }
            if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
              (**(code **)(plVar1[1] + 0x30))(plVar1);
            }
            if ((int)*plVar4 < 0) {
              return pcVar11;
            }
            *plVar4 = *plVar4 + -1;
            if (*plVar4 != 0) {
              return pcVar11;
            }
            (**(code **)(plVar4[1] + 0x30))(plVar4);
            return pcVar11;
          }
          plVar8 = (longlong *)FUN_180381a40(plVar3,plVar4);
          if (plVar8 != (longlong *)0x0) {
            plVar9 = (longlong *)PyNumber_Long(plVar8);
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            if (plVar9 != (longlong *)0x0) {
              pcVar11 = (code *)FUN_180383ba0(plVar2,plVar9);
              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                (**(code **)(plVar9[1] + 0x30))(plVar9);
              }
              if (pcVar11 != (code *)0x0) goto LAB_1800b60d0;
            }
          }
        }
      }
      lVar12 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    uVar13 = 0x2d;
    if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
      (**(code **)(plVar2[1] + 0x30))(plVar2);
    }
    if (((local_res20 != (code *)0x0) && (-1 < *(int *)local_res20)) &&
       (*(longlong *)local_res20 = *(longlong *)local_res20 + -1, *(longlong *)local_res20 == 0)) {
      (**(code **)(*(longlong *)(local_res20 + 8) + 0x30))();
    }
  }
  piVar5 = *(int **)(lVar12 + 0x28);
  if (piVar5 == (int *)0x0) {
    pcVar11 = (code *)FUN_18038e2a0(plVar6,uVar13);
    plVar8 = *(longlong **)(lVar12 + 0x28);
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar5 + 6) == plVar6) goto LAB_1800b5fe4;
    pcVar11 = (code *)FUN_18038e2a0(plVar6,uVar13);
    *(int **)(pcVar11 + 0x10) = piVar5;
    if (*piVar5 + 1 != 0) {
      *piVar5 = *piVar5 + 1;
    }
    plVar8 = *(longlong **)(lVar12 + 0x28);
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar10 = (code *)0x0;
    }
    else {
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  *(code **)(lVar12 + 0x28) = pcVar10;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))();
  }
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
  }
LAB_1800b5fe4:
  FUN_18036d0a0(plVar6,&DAT_180391c1c,plVar2,plVar3,plVar1,plVar4);
  plVar8 = DAT_1803b1f18;
  if (plVar6 == DAT_1803b1f18) {
    if ((-1 < (int)*DAT_1803b1f18) && (*DAT_1803b1f18 = *DAT_1803b1f18 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    DAT_1803b1f18 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))(plVar3);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar12;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}
