/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */
longlong * FUN_1800cabe0(longlong param_1,undefined8 param_2,undefined8 *param_3)
{
  int *piVar1;
  code *pcVar2;
  char cVar3;
  int iVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  undefined8 uVar11;
  undefined8 uVar12;
  longlong lVar13;
  longlong lVar14;
  undefined4 uVar15;
  longlong *plVar16;
  undefined1 auStack_228 [32];
  longlong *local_208;
  longlong *local_200;
  longlong *local_1f8;
  longlong *local_1f0;
  longlong *local_1e8;
  longlong *local_1e0;
  longlong *local_1d8;
  longlong *local_1d0;
  longlong *local_1c8;
  longlong *local_1c0;
  longlong *local_1b8;
  undefined8 local_1b0;
  longlong local_1a8;
  longlong *local_1a0;
  longlong *local_198;
  longlong *local_190;
  longlong *local_188;
  longlong *local_178;
  longlong *local_170;
  longlong *local_168;
  longlong *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  longlong *local_140;
  longlong *local_138;
  longlong *local_130;
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  longlong local_100;
  longlong *local_f8;
  longlong *local_f0;
  undefined8 local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  ulonglong local_48;
  plVar8 = DAT_1803b3618;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_228;
  local_138 = (longlong *)*param_3;
  plVar7 = (longlong *)0x0;
  plVar9 = (longlong *)param_3[1];
  plVar16 = (longlong *)0x0;
  local_1a8 = 0;
  local_130 = (longlong *)0x0;
  local_128 = (longlong *)0x0;
  local_120 = (longlong *)0x0;
  local_118 = (longlong *)0x0;
  local_110 = (longlong *)0x0;
  local_108 = (longlong *)0x0;
  local_148 = (longlong *)0x0;
  local_140 = (longlong *)0x0;
  local_f0 = (longlong *)0x0;
  local_1a0 = (longlong *)0x0;
  local_198 = (longlong *)0x0;
  local_158 = (longlong *)0x0;
  local_150 = (longlong *)0x0;
  local_190 = DAT_1803b3618;
  local_188 = plVar9;
  if (DAT_1803b3618 == (longlong *)0x0) {
LAB_1800cac9b:
    DAT_1803b3618 = (longlong *)FUN_18036cd10(param_1,DAT_1803b2a30,DAT_1803ca7a0,0x70);
  }
  else if ((1 < *DAT_1803b3618) || (DAT_1803b3618[2] != 0)) {
    if ((DAT_1803b3618 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3618 && (*DAT_1803b3618 = *DAT_1803b3618 + -1, *plVar8 == 0)))) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    goto LAB_1800cac9b;
  }
  plVar8 = DAT_1803b3618;
  local_190 = DAT_1803b3618;
  FUN_180001d00(param_1,DAT_1803b3618);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
  plVar5 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803b2b00);
  uVar15 = 0x6a;
  if (plVar5 == (longlong *)0x0) {
    local_1a8 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar16 = plVar7;
  }
  else {
    *(undefined4 *)(plVar8 + 5) = 0x6a;
    plVar6 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803b2c30);
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    if (plVar6 == (longlong *)0x0) {
      local_1a8 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar16 = plVar7;
    }
    else {
      iVar4 = FUN_180001510(plVar6);
      plVar7 = plVar16;
      if (iVar4 == -1) goto LAB_1800cad5f;
      if (iVar4 != 1) {
        if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
          (**(code **)(plVar6[1] + 0x30))(plVar6);
        }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
        plVar6 = DAT_1803b2b10;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
        iVar4 = (int)*DAT_1803b2b10 + 1;
        if (iVar4 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
          *(int *)DAT_1803b2b10 = iVar4;
        }
      }
      plVar5 = (longlong *)FUN_1803749a0(plVar6);
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      if (plVar5 == (longlong *)0x0) {
        local_1a8 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#3="strip" bytes=[3952481, 3952488] */
        plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar5,DAT_1803b2b18);
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
        if (plVar6 == (longlong *)0x0) {
          local_1a8 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          *(undefined4 *)(plVar8 + 5) = 0x6a;
          local_f8 = (longlong *)FUN_180376210(param_1,plVar6);
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          if (local_f8 == (longlong *)0x0) {
            local_1a8 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            local_130 = local_f8;
            iVar4 = FUN_180001510(local_f8);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#39="仿写短篇需要参考原文" bytes=[3953065, 3953097] */
            uVar12 = DAT_1803b2c38;
            pcVar2 = PyExc_ValueError_exref;
            if (iVar4 == -1) {
              local_1a8 = *(longlong *)(param_1 + 0x60);
              uVar15 = 0x6b;
              *(undefined8 *)(param_1 + 0x60) = 0;
            }
            else if (iVar4 == 0) {
              uVar15 = 0x6c;
              *(undefined4 *)(plVar8 + 5) = 0x6c;
              local_1a8 = FUN_180376770(param_1,*(undefined8 *)pcVar2,uVar12);
              FUN_18000d1d0(param_1,local_1a8);
            }
            else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
              plVar5 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803b2b00);
              uVar15 = 0x6e;
              if (plVar5 == (longlong *)0x0) {
                local_1a8 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
              }
              else {
                *(undefined4 *)(plVar8 + 5) = 0x6e;
                plVar6 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803b2c40);
                if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                  (**(code **)(plVar5[1] + 0x30))(plVar5);
                }
                if (plVar6 == (longlong *)0x0) {
                  local_1a8 = *(longlong *)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                else {
                  iVar4 = FUN_180001510(plVar6);
                  if (iVar4 == -1) {
LAB_1800cad5f:
                    local_1a8 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                    }
                  }
                  else {
                    if (iVar4 != 1) {
                      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                      }
                      plVar6 = DAT_1803b2c48;
                      iVar4 = (int)*DAT_1803b2c48 + 1;
                      if (iVar4 != 0) {
                        *(int *)DAT_1803b2c48 = iVar4;
                      }
                    }
                    plVar5 = (longlong *)PyNumber_Long(plVar6);
                    local_e0 = plVar5;
                    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                    }
                    if (plVar5 == (longlong *)0x0) {
                      local_1a8 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                    }
                    else {
                      local_128 = plVar5;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                      plVar5 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803b2b00);
                      uVar15 = 0x6f;
                      if (plVar5 == (longlong *)0x0) {
                        local_1a8 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                      }
                      else {
                        *(undefined4 *)(plVar8 + 5) = 0x6f;
                        plVar6 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803b2c50);
                        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                          (**(code **)(plVar5[1] + 0x30))(plVar5);
                        }
                        if (plVar6 == (longlong *)0x0) {
                          local_1a8 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                        }
                        else {
                          iVar4 = FUN_180001510(plVar6);
                          if (iVar4 == -1) goto LAB_1800cad5f;
                          if (iVar4 != 1) {
                            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                              (**(code **)(plVar6[1] + 0x30))(plVar6);
                            }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                            plVar6 = DAT_1803b2b10;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                            iVar4 = (int)*DAT_1803b2b10 + 1;
                            if (iVar4 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                              *(int *)DAT_1803b2b10 = iVar4;
                            }
                          }
                          plVar5 = (longlong *)FUN_1803749a0(plVar6);
                          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                          }
                          if (plVar5 == (longlong *)0x0) {
                            local_1a8 = *(longlong *)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#3="strip" bytes=[3952481, 3952488] */
                            plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar5,DAT_1803b2b18);
                            if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                              (**(code **)(plVar5[1] + 0x30))(plVar5);
                            }
                            if (plVar6 == (longlong *)0x0) {
                              local_1a8 = *(longlong *)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                            }
                            else {
                              *(undefined4 *)(plVar8 + 5) = 0x6f;
                              plVar5 = (longlong *)FUN_180376210(param_1,plVar6);
                              local_d8 = plVar5;
                              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                (**(code **)(plVar6[1] + 0x30))(plVar6);
                              }
                              if (plVar5 == (longlong *)0x0) {
                                local_1a8 = *(longlong *)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                              }
                              else {
                                local_120 = plVar5;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                plVar5 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803b2b00);
                                if (plVar5 == (longlong *)0x0) {
LAB_1800cb166:
                                  local_1a8 = *(longlong *)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                }
                                else {
                                  *(undefined4 *)(plVar8 + 5) = 0x70;
                                  plVar6 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803b2c58);
                                  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                  {
                                    (**(code **)(plVar5[1] + 0x30))(plVar5);
                                  }
                                  if (plVar6 == (longlong *)0x0) goto LAB_1800cb166;
                                  iVar4 = FUN_180001510(plVar6);
                                  if (iVar4 != -1) {
                                    if (iVar4 != 1) {
                                      if ((-1 < (int)*plVar6) &&
                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                                      }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                      plVar6 = DAT_1803b2b10;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                      iVar4 = (int)*DAT_1803b2b10 + 1;
                                      if (iVar4 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                        *(int *)DAT_1803b2b10 = iVar4;
                                      }
                                    }
                                    plVar5 = (longlong *)FUN_1803749a0(plVar6);
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                                    }
                                    if (plVar5 != (longlong *)0x0) {
                                      plVar6 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#3="strip" bytes=[3952481, 3952488] */
                                               FUN_18037f0c0(param_1,plVar5,DAT_1803b2b18);
                                      if ((-1 < (int)*plVar5) &&
                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                        (**(code **)(plVar5[1] + 0x30))(plVar5);
                                      }
                                      if (plVar6 != (longlong *)0x0) {
                                        *(undefined4 *)(plVar8 + 5) = 0x70;
                                        plVar5 = (longlong *)FUN_180376210(param_1,plVar6);
                                        local_d0 = plVar5;
                                        if ((-1 < (int)*plVar6) &&
                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                          (**(code **)(plVar6[1] + 0x30))(plVar6);
                                        }
                                        if (plVar5 != (longlong *)0x0) {
                                          local_118 = plVar5;
                                          plVar5 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                   FUN_18037f0c0(param_1,plVar9,DAT_1803b2b00);
                                          uVar15 = 0x71;
                                          if (plVar5 == (longlong *)0x0) {
                                            local_1a8 = *(longlong *)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                          }
                                          else {
                                            *(undefined4 *)(plVar8 + 5) = 0x71;
                                            plVar6 = (longlong *)
                                                     FUN_1803770d0(param_1,plVar5,DAT_1803b2c60);
                                            if ((-1 < (int)*plVar5) &&
                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                              (**(code **)(plVar5[1] + 0x30))(plVar5);
                                            }
                                            if (plVar6 == (longlong *)0x0) {
                                              local_1a8 = *(longlong *)(param_1 + 0x60);
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                            }
                                            else {
                                              iVar4 = FUN_180001510(plVar6);
                                              if (iVar4 == -1) goto LAB_1800cad5f;
                                              if (iVar4 != 1) {
                                                if ((-1 < (int)*plVar6) &&
                                                   (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                  lVar14 = FUN_18000cc40(plVar6);
                                                  (**(code **)(lVar14 + 0x30))(plVar6);
                                                }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                plVar6 = DAT_1803b2b10;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                iVar4 = (int)*DAT_1803b2b10 + 1;
                                                if (iVar4 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                  *(int *)DAT_1803b2b10 = iVar4;
                                                }
                                              }
                                              plVar5 = (longlong *)FUN_1803749a0(plVar6);
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar14 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar14 + 0x30))(plVar6);
                                              }
                                              if (plVar5 == (longlong *)0x0) {
                                                FUN_18000ce30(param_1,&local_1a8);
                                              }
                                              else {
                                                plVar6 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#3="strip" bytes=[3952481, 3952488] */
                                                         FUN_18037f0c0(param_1,plVar5,DAT_1803b2b18)
                                                ;
                                                if ((-1 < (int)*plVar5) &&
                                                   (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                  lVar14 = FUN_18000cc40(plVar5);
                                                  (**(code **)(lVar14 + 0x30))(plVar5);
                                                }
                                                if (plVar6 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_1a8);
                                                }
                                                else {
                                                  *(undefined4 *)(plVar8 + 5) = 0x71;
                                                  plVar5 = (longlong *)FUN_180376210(param_1,plVar6)
                                                  ;
                                                  local_c8 = plVar5;
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar14 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar14 + 0x30))(plVar6);
                                                  }
                                                  if (plVar5 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1a8);
                                                  }
                                                  else {
                                                    local_110 = plVar5;
                                                    plVar5 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar9,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                                           DAT_1803b2b00);
                                                    if (plVar5 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      uVar15 = 0x73;
                                                    }
                                                    else {
                                                      uVar15 = 0x73;
                                                      *(undefined4 *)(plVar8 + 5) = 0x73;
                                                      plVar6 = (longlong *)
                                                               FUN_1803770d0(param_1,plVar5,
                                                                             DAT_1803b2c68);
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar14 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar14 + 0x30))(plVar5);
                                                      }
                                                      if (plVar6 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                      }
                                                      else {
                                                        iVar4 = FUN_180001510(plVar6);
                                                        if (iVar4 == -1) {
LAB_1800cb4c9:
                                                          FUN_18000ce30(param_1,&local_1a8);
                                                          if ((-1 < (int)*plVar6) &&
                                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar6);
                                                            (**(code **)(lVar14 + 0x30))(plVar6);
                                                          }
                                                        }
                                                        else if (iVar4 == 1) {
LAB_1800cb5d8:
                                                          plVar5 = (longlong *)FUN_1803749a0(plVar6)
                                                          ;
                                                          if ((-1 < (int)*plVar6) &&
                                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                                          {
                                                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                          }
                                                          if (plVar5 == (longlong *)0x0) {
                                                            local_1a8 = *(longlong *)
                                                                         (param_1 + 0x60);
                                                            uVar15 = 0x72;
                                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                                          }
                                                          else {
                                                            plVar6 = (longlong *)
                                                                     FUN_18037f0c0(param_1,plVar5,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#3="strip" bytes=[3952481, 3952488] */
                                                                                   DAT_1803b2b18);
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              (**(code **)(plVar5[1] + 0x30))
                                                                        (plVar5);
                                                            }
                                                            uVar15 = 0x72;
                                                            if (plVar6 == (longlong *)0x0) {
                                                              local_1a8 = *(longlong *)
                                                                           (param_1 + 0x60);
                                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                                            }
                                                            else {
                                                              *(undefined4 *)(plVar8 + 5) = 0x72;
                                                              plVar5 = (longlong *)
                                                                       FUN_180376210(param_1,plVar6)
                                                              ;
                                                              local_178 = plVar5;
                                                              if ((-1 < (int)*plVar6) &&
                                                                 (*plVar6 = *plVar6 + -1,
                                                                 *plVar6 == 0)) {
                                                                (**(code **)(plVar6[1] + 0x30))
                                                                          (plVar6);
                                                              }
                                                              if (plVar5 == (longlong *)0x0) {
                                                                local_1a8 = *(longlong *)
                                                                             (param_1 + 0x60);
                                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                              }
                                                              else {
                                                                local_108 = plVar5;
                                                                lVar14 = FUN_1800c7720(param_1);
                                                                if (lVar14 == 0) {
                                                                  FUN_180373340(param_1,&local_1a8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#47="resolve_imitation_short_story_id" bytes=[3953206, 3953240] */
                                                                                DAT_1803b2c78);
                                                                  uVar15 = 0x76;
                                                                }
                                                                else {
                                                                  uVar15 = 0x76;
                                                                  *(undefined4 *)(plVar8 + 5) = 0x76
                                                                  ;
                                                                  plVar6 = (longlong *)
                                                                           FUN_180376770(param_1,
                                                  lVar14,plVar9);
                                                  plVar5 = local_138;
                                                  local_168 = plVar6;
                                                  if (plVar6 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1a8);
                                                  }
                                                  else {
                                                    local_148 = plVar6;
                                                    cVar3 = FUN_18037f3a0(param_1,local_138,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#48="_current_story_id" bytes=[3953240, 3953259] */
                                                                          DAT_1803b2c80,plVar6);
                                                    if (cVar3 == '\0') {
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      uVar15 = 0x77;
                                                    }
                                                    else {
                                                      plVar5 = (longlong *)
                                                               FUN_18037f0c0(param_1,plVar5,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#14="shorts_root" bytes=[3952645, 3952658] */
                                                                             DAT_1803b2b70);
                                                      if (plVar5 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        uVar15 = 0x78;
                                                      }
                                                      else {
                                                        plVar7 = (longlong *)
                                                                 FUN_180382540(plVar5,plVar6);
                                                        local_160 = plVar7;
                                                        if ((-1 < (int)*plVar5) &&
                                                           (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                          lVar14 = FUN_18000cc40(plVar5);
                                                          (**(code **)(lVar14 + 0x30))(plVar5);
                                                        }
                                                        if (plVar7 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1a8);
                                                          uVar15 = 0x78;
                                                          plVar7 = local_198;
                                                          plVar16 = local_198;
                                                        }
                                                        else {
                                                          local_140 = plVar7;
                                                          lVar14 = FUN_1800c7020(param_1);
                                                          if (lVar14 == 0) {
                                                            FUN_180373340(param_1,&local_1a8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                                          DAT_1803b2b20);
                                                            uVar15 = 0x79;
                                                            plVar7 = local_198;
                                                            plVar16 = local_198;
                                                          }
                                                          else {
                                                            plVar16 = (longlong *)
                                                                      FUN_18037f0c0(param_1,lVar14,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#49="ensure_dir" bytes=[3953259, 3953271] */
                                                                                    DAT_1803b2c88);
                                                            if (plVar16 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1a8);
                                                              uVar15 = 0x79;
                                                              plVar7 = local_198;
                                                              plVar16 = local_198;
                                                            }
                                                            else {
                                                              plVar5 = (longlong *)
                                                                       FUN_180382540(plVar7,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#50="imitation" bytes=[3953271, 3953282] */
                                                  DAT_1803b2c90);
                                                  if (plVar5 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1a8);
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar16);
                                                      (**(code **)(lVar14 + 0x30))(plVar16);
                                                    }
                                                    uVar15 = 0x79;
                                                    plVar7 = local_198;
                                                    plVar16 = local_198;
                                                  }
                                                  else {
                                                    uVar15 = 0x79;
                                                    *(undefined4 *)(plVar8 + 5) = 0x79;
                                                    plVar8 = (longlong *)
                                                             FUN_180376770(param_1,plVar16,plVar5);
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar16);
                                                      (**(code **)(lVar14 + 0x30))(plVar16);
                                                    }
                                                    if ((-1 < (int)*plVar5) &&
                                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar5);
                                                      (**(code **)(lVar14 + 0x30))(plVar5);
                                                    }
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      plVar7 = local_198;
                                                      plVar16 = local_198;
                                                      plVar8 = local_190;
                                                    }
                                                    else {
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar14 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar14 + 0x30))(plVar8);
                                                      }
                                                      lVar14 = FUN_1800c7020(param_1);
                                                      if (lVar14 == 0) {
                                                        FUN_180373340(param_1,&local_1a8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                                      DAT_1803b2b20);
                                                        uVar15 = 0x7a;
                                                        plVar7 = local_198;
                                                        plVar16 = local_198;
                                                        plVar8 = local_190;
                                                      }
                                                      else {
                                                        plVar16 = (longlong *)
                                                                  FUN_18037f0c0(param_1,lVar14,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#49="ensure_dir" bytes=[3953259, 3953271] */
                                                                                DAT_1803b2c88);
                                                        if (plVar16 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1a8);
                                                          uVar15 = 0x7a;
                                                          plVar7 = local_198;
                                                          plVar16 = local_198;
                                                          plVar8 = local_190;
                                                        }
                                                        else {
                                                          plVar5 = (longlong *)
                                                                   FUN_180382540(plVar7,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#51="outline" bytes=[3953282, 3953291] */
                                                  DAT_1803b2c98);
                                                  plVar8 = local_190;
                                                  if (plVar5 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1a8);
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar16);
                                                      (**(code **)(lVar14 + 0x30))(plVar16);
                                                    }
                                                    uVar15 = 0x7a;
                                                    plVar7 = local_198;
                                                    plVar16 = local_198;
                                                    plVar8 = local_190;
                                                  }
                                                  else {
                                                    uVar15 = 0x7a;
                                                    *(undefined4 *)(local_190 + 5) = 0x7a;
                                                    plVar9 = (longlong *)
                                                             FUN_180376770(param_1,plVar16,plVar5);
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar16);
                                                      (**(code **)(lVar14 + 0x30))(plVar16);
                                                    }
                                                    if ((-1 < (int)*plVar5) &&
                                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar5);
                                                      (**(code **)(lVar14 + 0x30))(plVar5);
                                                    }
                                                    if (plVar9 == (longlong *)0x0) {
LAB_1800cb9c9:
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_198;
                                                    }
                                                    else {
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar14 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar14 + 0x30))(plVar9);
                                                      }
                                                      lVar14 = FUN_1800c7020(param_1);
                                                      if (lVar14 == 0) {
                                                        FUN_180373340(param_1,&local_1a8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                                      DAT_1803b2b20);
                                                        uVar15 = 0x7b;
                                                        plVar9 = local_188;
                                                        plVar7 = local_198;
                                                        plVar16 = local_198;
                                                      }
                                                      else {
                                                        plVar9 = (longlong *)
                                                                 FUN_18037f0c0(param_1,lVar14,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#49="ensure_dir" bytes=[3953259, 3953271] */
                                                                               DAT_1803b2c88);
                                                        if (plVar9 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1a8);
                                                          uVar15 = 0x7b;
                                                          plVar9 = local_188;
                                                          plVar7 = local_198;
                                                          plVar16 = local_198;
                                                        }
                                                        else {
                                                          plVar16 = (longlong *)
                                                                    FUN_180382540(plVar7,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#52="reviews" bytes=[3953291, 3953300] */
                                                  DAT_1803b2ca0);
                                                  if (plVar16 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1a8);
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar14 + 0x30))(plVar9);
                                                    }
                                                    uVar15 = 0x7b;
                                                    plVar9 = local_188;
                                                    plVar7 = local_198;
                                                    plVar16 = local_198;
                                                  }
                                                  else {
                                                    uVar15 = 0x7b;
                                                    *(undefined4 *)(plVar8 + 5) = 0x7b;
                                                    plVar5 = (longlong *)
                                                             FUN_180376770(param_1,plVar9,plVar16);
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar14 + 0x30))(plVar9);
                                                    }
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar16);
                                                      (**(code **)(lVar14 + 0x30))(plVar16);
                                                    }
                                                    if (plVar5 == (longlong *)0x0)
                                                    goto LAB_1800cb9c9;
                                                    if ((-1 < (int)*plVar5) &&
                                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar5);
                                                      (**(code **)(lVar14 + 0x30))(plVar5);
                                                    }
                                                    lVar14 = FUN_1800c7020(param_1);
                                                    if (lVar14 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                      FUN_180373340(param_1,&local_1a8,DAT_1803b2b20
                                                                   );
                                                      uVar15 = 0x7c;
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_198;
                                                    }
                                                    else {
                                                      plVar9 = (longlong *)
                                                               FUN_18037f0c0(param_1,lVar14,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#49="ensure_dir" bytes=[3953259, 3953271] */
                                                                             DAT_1803b2c88);
                                                      if (plVar9 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        uVar15 = 0x7c;
                                                        plVar9 = local_188;
                                                        plVar7 = local_198;
                                                        plVar16 = local_198;
                                                      }
                                                      else {
                                                        plVar16 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#53="drafts" bytes=[3953300, 3953308] */
                                                                  FUN_180382540(plVar7,DAT_1803b2ca8
                                                                               );
                                                        if (plVar16 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1a8);
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar14 + 0x30))(plVar9);
                                                          }
                                                          uVar15 = 0x7c;
                                                          plVar9 = local_188;
                                                          plVar7 = local_198;
                                                          plVar16 = local_198;
                                                        }
                                                        else {
                                                          uVar15 = 0x7c;
                                                          *(undefined4 *)(plVar8 + 5) = 0x7c;
                                                          plVar5 = (longlong *)
                                                                   FUN_180376770(param_1,plVar9,
                                                                                 plVar16);
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar14 + 0x30))(plVar9);
                                                          }
                                                          if ((-1 < (int)*plVar16) &&
                                                             (*plVar16 = *plVar16 + -1,
                                                             *plVar16 == 0)) {
                                                            lVar14 = FUN_18000cc40(plVar16);
                                                            (**(code **)(lVar14 + 0x30))(plVar16);
                                                          }
                                                          if (plVar5 == (longlong *)0x0)
                                                          goto LAB_1800cb9c9;
                                                          if ((-1 < (int)*plVar5) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar5);
                                                            (**(code **)(lVar14 + 0x30))(plVar5);
                                                          }
                                                          lVar14 = FUN_1800c7020(param_1);
                                                          if (lVar14 == 0) {
                                                            FUN_180373340(param_1,&local_1a8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                                          DAT_1803b2b20);
                                                            uVar15 = 0x7d;
                                                            plVar9 = local_188;
                                                            plVar7 = local_198;
                                                            plVar16 = local_198;
                                                          }
                                                          else {
                                                            plVar9 = (longlong *)
                                                                     FUN_18037f0c0(param_1,lVar14,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#49="ensure_dir" bytes=[3953259, 3953271] */
                                                                                   DAT_1803b2c88);
                                                            if (plVar9 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1a8);
                                                              uVar15 = 0x7d;
                                                              plVar9 = local_188;
                                                              plVar7 = local_198;
                                                              plVar16 = local_198;
                                                            }
                                                            else {
                                                              plVar16 = (longlong *)
                                                                        FUN_180382540(plVar7,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#54="final" bytes=[3953308, 3953315] */
                                                  DAT_1803b2cb0);
                                                  if (plVar16 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1a8);
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar14 + 0x30))(plVar9);
                                                    }
                                                    uVar15 = 0x7d;
                                                    plVar9 = local_188;
                                                    plVar7 = local_198;
                                                    plVar16 = local_198;
                                                  }
                                                  else {
                                                    uVar15 = 0x7d;
                                                    *(undefined4 *)(plVar8 + 5) = 0x7d;
                                                    plVar5 = (longlong *)
                                                             FUN_180376770(param_1,plVar9,plVar16);
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar14 + 0x30))(plVar9);
                                                    }
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar16);
                                                      (**(code **)(lVar14 + 0x30))(plVar16);
                                                    }
                                                    if (plVar5 == (longlong *)0x0)
                                                    goto LAB_1800cb9c9;
                                                    if ((-1 < (int)*plVar5) &&
                                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar5);
                                                      (**(code **)(lVar14 + 0x30))(plVar5);
                                                    }
                                                    plVar9 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#54="final" bytes=[3953308, 3953315] */
                                                             FUN_180382540(plVar7,DAT_1803b2cb0);
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      uVar15 = 0x80;
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_198;
                                                    }
                                                    else {
                                                      plVar16 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#55="full.md" bytes=[3953315, 3953324] */
                                                                FUN_180382540(plVar9,DAT_1803b2cb8);
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar14 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar14 + 0x30))(plVar9);
                                                      }
                                                      if (plVar16 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        uVar15 = 0x80;
                                                        plVar9 = local_188;
                                                        plVar7 = local_198;
                                                        plVar16 = local_198;
                                                      }
                                                      else {
                                                        uVar15 = 0x80;
                                                        *(undefined4 *)(plVar8 + 5) = 0x80;
                                                        plVar5 = (longlong *)
                                                                 FUN_18037e560(param_1,plVar16,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#56="exists" bytes=[3953324, 3953332] */
                                                                               DAT_1803b2cc0);
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar14 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar14 + 0x30))(plVar16);
                                                        }
                                                        if (plVar5 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1a8);
                                                          plVar9 = local_188;
                                                          plVar7 = local_198;
                                                          plVar16 = local_198;
                                                        }
                                                        else {
                                                          iVar4 = FUN_180001510(plVar5);
                                                          if (iVar4 == -1) {
                                                            FUN_18000ce30(param_1,&local_1a8);
                                                            plVar9 = local_188;
                                                            plVar7 = local_198;
                                                            plVar16 = local_198;
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar14 + 0x30))(plVar5);
                                                              plVar9 = local_188;
                                                              plVar7 = local_198;
                                                              plVar16 = local_198;
                                                            }
                                                          }
                                                          else {
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar14 + 0x30))(plVar5);
                                                            }
                                                            if (iVar4 == 0) {
LAB_1800cc5f7:
                                                              plVar7 = (longlong *)
                                                                       FUN_18037f0c0(param_1,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#18="_emit" bytes=[3952732, 3952739] */
                                                  local_138,DAT_1803b2b90);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#69="imitation_short_run" bytes=[3953458, 3953479] */
                                                  plVar9 = DAT_1803b2d28;
                                                  if (plVar7 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_1a8);
                                                    uVar15 = 0x8c;
                                                    plVar9 = local_188;
                                                    plVar7 = local_198;
                                                    plVar16 = local_198;
                                                  }
                                                  else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#69="imitation_short_run" bytes=[3953458, 3953479] */
                                                    local_b0 = DAT_1803b2d28;
                                                    lVar14 = FUN_1800c7130(param_1);
                                                    if (lVar14 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#70="IMIT_SHORT_STAGES" bytes=[3953479, 3953498] */
                                                      FUN_180373340(param_1,&local_1a8,DAT_1803b2d30
                                                                   );
                                                      lVar14 = local_1a8;
                                                      FUN_18000ce70(&local_1a8);
                                                    }
                                                    else {
                                                      plVar8 = (longlong *)
                                                               FUN_18000d280(param_1,lVar14);
                                                      if (plVar8 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        lVar14 = local_1a8;
                                                        FUN_18000ce70(&local_1a8);
                                                      }
                                                      else {
                                                        plVar10 = (longlong *)
                                                                  FUN_180372160(param_1,0);
                                                        plVar16 = (longlong *)FUN_18000d430(plVar8);
                                                        plVar5 = local_158;
                                                        plVar6 = local_150;
                                                        while (local_158 = plVar5,
                                                              plVar16 != (longlong *)0x0) {
                                                          local_150 = plVar16;
                                                          if (((plVar6 != (longlong *)0x0) &&
                                                              (-1 < (int)*plVar6)) &&
                                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar6);
                                                            (**(code **)(lVar14 + 0x30))(plVar6);
                                                          }
                                                          iVar4 = (int)*plVar16 + 1;
                                                          if (iVar4 != 0) {
                                                            *(int *)plVar16 = iVar4;
                                                          }
                                                          local_158 = plVar16;
                                                          if (((plVar5 != (longlong *)0x0) &&
                                                              (-1 < (int)*plVar5)) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40();
                                                            (**(code **)(lVar14 + 0x30))(plVar5);
                                                          }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#71="name" bytes=[3953498, 3953504] */
                                                          uVar12 = DAT_1803b2d38;
                                                          uVar11 = _PyDict_NewPresized(2);
                                                          PyDict_SetItem(uVar11,uVar12,plVar16);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#72="status" bytes=[3953504, 3953512] */
                                                          PyDict_SetItem(uVar11,DAT_1803b2d40,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#73="pending" bytes=[3953512, 3953521] */
                                                                         DAT_1803b2d48);
                                                          cVar3 = FUN_180372560(plVar10,uVar11);
                                                          if ((cVar3 == '\0') ||
                                                             (cVar3 = FUN_18000d5b0(param_1),
                                                             cVar3 == '\0')) goto LAB_1800ccf64;
                                                          plVar16 = (longlong *)
                                                                    FUN_18000d430(plVar8);
                                                          plVar9 = local_b0;
                                                          plVar5 = local_158;
                                                          plVar6 = local_150;
                                                        }
                                                        local_150 = plVar6;
                                                        cVar3 = FUN_18000cda0(param_1);
                                                        if (cVar3 != '\0') {
                                                          iVar4 = (int)*plVar10 + 1;
                                                          if (iVar4 != 0) {
                                                            *(int *)plVar10 = iVar4;
                                                          }
                                                          if ((-1 < (int)*plVar8) &&
                                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar8);
                                                            (**(code **)(lVar14 + 0x30))(plVar8);
                                                          }
                                                          if ((-1 < (int)*plVar10) &&
                                                             (*plVar10 = *plVar10 + -1,
                                                             *plVar10 == 0)) {
                                                            lVar14 = FUN_18000cc40(plVar10);
                                                            (**(code **)(lVar14 + 0x30))(plVar10);
                                                          }
                                                          plVar8 = local_158;
                                                          if ((local_158 != (longlong *)0x0) &&
                                                             (-1 < (int)*local_158)) {
                                                            *local_158 = *local_158 + -1;
                                                            iVar4 = (int)*local_158;
                                                            if (*local_158 == 0) {
                                                              lVar14 = FUN_18000cc40(local_158);
                                                              (**(code **)(lVar14 + 0x30))(plVar8);
                                                              iVar4 = (int)*plVar8;
                                                            }
                                                            if ((-1 < iVar4) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar14 + 0x30))(plVar8);
                                                            }
                                                          }
                                                          plVar16 = local_168;
                                                          uVar15 = 0x8c;
                                                          local_a8 = local_168;
                                                          *(undefined4 *)(local_190 + 5) = 0x8c;
                                                          local_208 = (longlong *)DAT_1803b2d58;
                                                          local_b0 = plVar9;
                                                          local_a0 = plVar10;
                                                          plVar9 = (longlong *)
                                                                   FUN_18037d280(param_1,plVar7,
                                                                                 DAT_1803b2d50,
                                                                                 &local_b0);
                                                          if ((-1 < (int)*plVar7) &&
                                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar7);
                                                            (**(code **)(lVar14 + 0x30))(plVar7);
                                                          }
                                                          if ((-1 < (int)*plVar10) &&
                                                             (*plVar10 = *plVar10 + -1,
                                                             *plVar10 == 0)) {
                                                            lVar14 = FUN_18000cc40(plVar10);
                                                            (**(code **)(lVar14 + 0x30))(plVar10);
                                                          }
                                                          if (plVar9 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_1a8);
                                                            plVar9 = local_188;
                                                            plVar7 = local_198;
                                                            plVar16 = local_198;
                                                            plVar8 = local_190;
                                                          }
                                                          else {
                                                            if ((-1 < (int)*plVar9) &&
                                                               (*plVar9 = *plVar9 + -1, *plVar9 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar9);
                                                              (**(code **)(lVar14 + 0x30))(plVar9);
                                                            }
                                                            plVar7 = (longlong *)
                                                                     FUN_18037f0c0(param_1,local_138
/* Analyst constant binding: novels.pipeline.imitation_short_runner#76="_produce" bytes=[3953560, 3953570] */
                                                                                   ,DAT_1803b2d60);
                                                            plVar9 = local_188;
                                                            plVar8 = local_190;
                                                            if (plVar7 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1a8);
                                                              plVar8 = local_190;
                                                            }
                                                            else {
                                                              local_98 = local_130;
                                                              local_90 = local_128;
                                                              local_88 = local_120;
                                                              local_80 = local_118;
                                                              local_78 = local_110;
                                                              local_70 = local_108;
                                                              *(undefined4 *)(local_190 + 5) = 0x94;
                                                              local_60 = local_140;
                                                              local_68 = plVar16;
                                                              local_58 = local_188;
                                                              plVar5 = (longlong *)
                                                                       FUN_18037ccb0(param_1,plVar7,
                                                                                     &local_98,
                                                                                     DAT_1803b2d68);
                                                              if ((-1 < (int)*plVar7) &&
                                                                 (*plVar7 = *plVar7 + -1,
                                                                 *plVar7 == 0)) {
                                                                lVar14 = FUN_18000cc40(plVar7);
                                                                (**(code **)(lVar14 + 0x30))(plVar7)
                                                                ;
                                                              }
                                                              plVar7 = local_160;
                                                              if (plVar5 != (longlong *)0x0)
                                                              goto LAB_1800ccd98;
                                                              FUN_18000ce30(param_1,&local_1a8);
                                                            }
                                                            local_100 = local_1a8;
                                                            FUN_18000ce70(&local_1a8);
                                                            local_e8 = FUN_18000ccc0(param_1);
                                                            lVar14 = FUN_18000ce80(&local_100);
                                                            if (lVar14 == 0) {
                                                              uVar12 = FUN_18038e2a0(plVar8,0x94);
                                                            }
                                                            else {
                                                              uVar12 = FUN_18000cc90(lVar14,plVar8,
                                                                                     0x94);
                                                            }
                                                            FUN_180001300(&local_100,uVar12);
                                                            FUN_18000d050(param_1,&local_100);
                                                            uVar12 = FUN_18000cc40(*(undefined8 *)
                                                                                    (param_1 + 0xf8)
                                                                                  );
                                                            iVar4 = FUN_18000ce90(param_1,uVar12,
                                                                                  *(undefined8 *)
                                                  PyExc_Exception_exref);
                                                  if (iVar4 == 0) {
                                                    cVar3 = FUN_180373290(param_1,&local_1a8);
                                                    uVar15 = 0x93;
                                                    if (cVar3 != '\0') {
                                                      uVar15 = 0;
                                                    }
                                                    lVar13 = FUN_18000ce80(&local_1a8);
                                                    lVar14 = local_1a8;
                                                    if ((lVar13 != 0) &&
                                                       (*(longlong **)(lVar13 + 0x18) == plVar8)) {
                                                      *(undefined4 *)(plVar8 + 5) =
                                                           *(undefined4 *)(lVar13 + 0x24);
                                                    }
                                                  }
                                                  else {
                                                    plVar9 = *(longlong **)(param_1 + 0xf8);
                                                    iVar4 = (int)*plVar9 + 1;
                                                    if (iVar4 != 0) {
                                                      *(int *)plVar9 = iVar4;
                                                    }
                                                    lVar14 = FUN_1800c7020(param_1);
                                                    if (lVar14 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                      FUN_180373340(param_1,&local_1a8,DAT_1803b2b20
                                                                   );
                                                      uVar15 = 0xa0;
                                                    }
                                                    else {
                                                      plVar7 = (longlong *)
                                                               FUN_18037f0c0(param_1,lVar14,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#78="write_json" bytes=[3953686, 3953698] */
                                                                             DAT_1803b2d70);
                                                      if (plVar7 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        uVar15 = 0xa0;
                                                      }
                                                      else {
                                                        plVar16 = (longlong *)
                                                                  FUN_180382540(local_140,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#58="status.json" bytes=[3953343, 3953356] */
                                                                                DAT_1803b2cd0);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#79="failed" bytes=[3953698, 3953706] */
                                                        uVar11 = DAT_1803b2d78;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#72="status" bytes=[3953504, 3953512] */
                                                        uVar12 = DAT_1803b2d40;
                                                        if (plVar16 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1a8);
                                                          if ((-1 < (int)*plVar7) &&
                                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar7);
                                                            (**(code **)(lVar14 + 0x30))(plVar7);
                                                          }
                                                          uVar15 = 0xa0;
                                                        }
                                                        else {
                                                          plVar5 = (longlong *)
                                                                   _PyDict_NewPresized(3);
                                                          PyDict_SetItem(plVar5,uVar12,uVar11);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#65="storyId" bytes=[3953414, 3953423] */
                                                          PyDict_SetItem(plVar5,DAT_1803b2d08,
                                                                         local_148);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#80="error" bytes=[3953706, 3953713] */
                                                          uVar12 = DAT_1803b2d80;
                                                          plVar8 = (longlong *)FUN_1803749a0(plVar9)
                                                          ;
                                                          if (plVar8 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_1a8);
                                                            uVar15 = 0xa3;
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar14 + 0x30))(plVar7);
                                                            }
                                                            if ((-1 < (int)*plVar16) &&
                                                               (*plVar16 = *plVar16 + -1,
                                                               *plVar16 == 0)) {
                                                              lVar14 = FUN_18000cc40(plVar16);
                                                              (**(code **)(lVar14 + 0x30))(plVar16);
                                                            }
                                                            plVar8 = local_190;
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar14 + 0x30))(plVar5);
                                                              plVar8 = local_190;
                                                            }
                                                          }
                                                          else {
                                                            PyDict_SetItem(plVar5,uVar12,plVar8);
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar14 + 0x30))(plVar8);
                                                            }
                                                            uVar15 = 0xa0;
                                                            *(undefined4 *)(local_190 + 5) = 0xa0;
                                                            local_178 = plVar16;
                                                            local_170 = plVar5;
                                                            plVar8 = (longlong *)
                                                                     FUN_180377750(param_1,plVar7,
                                                                                   &local_178);
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar7);
                                                              (**(code **)(lVar14 + 0x30))(plVar7);
                                                            }
                                                            if ((-1 < (int)*plVar16) &&
                                                               (*plVar16 = *plVar16 + -1,
                                                               *plVar16 == 0)) {
                                                              lVar14 = FUN_18000cc40(plVar16);
                                                              (**(code **)(lVar14 + 0x30))(plVar16);
                                                            }
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar14 + 0x30))(plVar5);
                                                            }
                                                            if (plVar8 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_1a8);
                                                              plVar8 = local_190;
                                                            }
                                                            else {
                                                              if ((-1 < (int)*plVar8) &&
                                                                 (*plVar8 = *plVar8 + -1,
                                                                 *plVar8 == 0)) {
                                                                lVar14 = FUN_18000cc40(plVar8);
                                                                (**(code **)(lVar14 + 0x30))(plVar8)
                                                                ;
                                                              }
                                                              cVar3 = FUN_180373290(param_1,&
                                                  local_1a8);
                                                  uVar15 = 0xa5;
                                                  if (cVar3 != '\0') {
                                                    uVar15 = 0;
                                                  }
                                                  lVar14 = FUN_18000ce80(&local_1a8);
                                                  plVar8 = local_190;
                                                  if ((lVar14 != 0) &&
                                                     (*(longlong **)(lVar14 + 0x18) == local_190)) {
                                                    *(undefined4 *)(local_190 + 5) =
                                                         *(undefined4 *)(lVar14 + 0x24);
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  lVar14 = local_1a8;
                                                  FUN_18000ce70(&local_1a8);
                                                  if (((plVar9 != (longlong *)0x0) &&
                                                      (-1 < (int)*plVar9)) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar13 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar13 + 0x30))(plVar9);
                                                  }
                                                  }
                                                  local_1a8 = lVar14;
                                                  lVar14 = local_1a8;
                                                  FUN_18000ce70(&local_1a8);
                                                  FUN_18000cce0(param_1,&local_e8);
                                                  plVar9 = local_188;
                                                  plVar7 = local_198;
                                                  plVar16 = local_198;
                                                  local_1a8 = lVar14;
                                                  }
                                                  goto LAB_1800cd021;
                                                  }
LAB_1800ccf64:
                                                  FUN_18000ce30(param_1,&local_1a8);
                                                  lVar14 = local_1a8;
                                                  FUN_18000ce70(&local_1a8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar13 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar13 + 0x30))(plVar8);
                                                  }
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    lVar13 = FUN_18000cc40(plVar10);
                                                    (**(code **)(lVar13 + 0x30))(plVar10);
                                                  }
                                                  plVar9 = local_158;
                                                  if (((local_158 != (longlong *)0x0) &&
                                                      (-1 < (int)*local_158)) &&
                                                     (*local_158 = *local_158 + -1, *local_158 == 0)
                                                     ) {
                                                    lVar13 = FUN_18000cc40(local_158);
                                                    (**(code **)(lVar13 + 0x30))(plVar9);
                                                  }
                                                  local_1a8 = lVar14;
                                                  FUN_18000ce70(&local_1a8);
                                                  if (((plVar9 != (longlong *)0x0) &&
                                                      (-1 < (int)*plVar9)) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar13 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar13 + 0x30))(plVar9);
                                                  }
                                                  }
                                                  }
                                                  uVar15 = 0x90;
                                                  plVar9 = local_188;
                                                  plVar7 = local_198;
                                                  plVar16 = local_198;
                                                  plVar8 = local_190;
                                                  local_1a8 = lVar14;
                                                  }
                                                  }
                                                  else {
                                                    lVar14 = FUN_1800c7020(param_1);
                                                    if (lVar14 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                      FUN_180373340(param_1,&local_1a8,DAT_1803b2b20
                                                                   );
                                                      uVar15 = 0x81;
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_198;
                                                    }
                                                    else {
                                                      plVar9 = (longlong *)
                                                               FUN_18037f0c0(param_1,lVar14,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#57="read_json" bytes=[3953332, 3953343] */
                                                                             DAT_1803b2cc8);
                                                      if (plVar9 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        uVar15 = 0x81;
                                                        plVar9 = local_188;
                                                        plVar7 = local_198;
                                                        plVar16 = local_198;
                                                      }
                                                      else {
                                                        plVar16 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#58="status.json" bytes=[3953343, 3953356] */
                                                                  FUN_180382540(plVar7,DAT_1803b2cd0
                                                                               );
                                                        if (plVar16 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_1a8);
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar14 + 0x30))(plVar9);
                                                          }
                                                          uVar15 = 0x81;
                                                          plVar9 = local_188;
                                                          plVar7 = local_198;
                                                          plVar16 = local_198;
                                                        }
                                                        else {
                                                          plVar5 = (longlong *)
                                                                   FUN_180370ec0(param_1);
                                                          uVar15 = 0x81;
                                                          *(undefined4 *)(plVar8 + 5) = 0x81;
                                                          local_c0 = plVar16;
                                                          local_b8 = plVar5;
                                                          plVar8 = (longlong *)
                                                                   FUN_180377750(param_1,plVar9,
                                                                                 &local_c0);
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar14 + 0x30))(plVar9);
                                                          }
                                                          if ((-1 < (int)*plVar16) &&
                                                             (*plVar16 = *plVar16 + -1,
                                                             *plVar16 == 0)) {
                                                            lVar14 = FUN_18000cc40(plVar16);
                                                            (**(code **)(lVar14 + 0x30))(plVar16);
                                                          }
                                                          if ((-1 < (int)*plVar5) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar5);
                                                            (**(code **)(lVar14 + 0x30))(plVar5);
                                                          }
                                                          if (plVar8 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_1a8);
                                                            plVar9 = local_188;
                                                            plVar7 = local_198;
                                                            plVar16 = local_198;
                                                            plVar8 = local_190;
                                                          }
                                                          else {
                                                            local_f0 = plVar8;
                                                            iVar4 = PyObject_IsInstance(plVar8,
                                                  PyDict_Type_exref);
                                                  if (iVar4 == -1) {
                                                    FUN_18000ce30(param_1,&local_1a8);
                                                    uVar15 = 0x82;
                                                    plVar9 = local_188;
                                                    plVar7 = local_198;
                                                    plVar16 = local_198;
                                                    plVar8 = local_190;
                                                  }
                                                  else if (iVar4 == 0) {
                                                    iVar4 = 0;
                                                    plVar8 = local_190;
LAB_1800cc11d:
                                                    if (iVar4 != 1) goto LAB_1800cc5f7;
                                                    lVar14 = FUN_1800c7020(param_1);
                                                    if (lVar14 == 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#4="FileRepository" bytes=[3952488, 3952504] */
                                                      FUN_180373340(param_1,&local_1a8,DAT_1803b2b20
                                                                   );
                                                      uVar15 = 0x83;
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_198;
                                                    }
                                                    else {
                                                      plVar9 = (longlong *)
                                                               FUN_18037f0c0(param_1,lVar14,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#61="read_text" bytes=[3953377, 3953388] */
                                                                             DAT_1803b2ce8);
                                                      if (plVar9 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        uVar15 = 0x83;
                                                        plVar9 = local_188;
                                                        plVar7 = local_198;
                                                        plVar16 = local_198;
                                                      }
                                                      else {
                                                        plVar16 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#54="final" bytes=[3953308, 3953315] */
                                                                  FUN_180382540(plVar7,DAT_1803b2cb0
                                                                               );
                                                        if (plVar16 != (longlong *)0x0) {
                                                          plVar5 = (longlong *)
                                                                   FUN_180382540(plVar16,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#55="full.md" bytes=[3953315, 3953324] */
                                                  DAT_1803b2cb8);
                                                  if ((-1 < (int)*plVar16) &&
                                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                    lVar14 = FUN_18000cc40(plVar16);
                                                    (**(code **)(lVar14 + 0x30))(plVar16);
                                                  }
                                                  if (plVar5 != (longlong *)0x0) {
                                                    uVar15 = 0x83;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                    local_a8 = DAT_1803b2b10;
                                                    *(undefined4 *)(plVar8 + 5) = 0x83;
                                                    local_b0 = plVar5;
                                                    plVar16 = (longlong *)
                                                              FUN_180377750(param_1,plVar9,&local_b0
                                                                           );
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar14 + 0x30))(plVar9);
                                                    }
                                                    if ((-1 < (int)*plVar5) &&
                                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar5);
                                                      (**(code **)(lVar14 + 0x30))(plVar5);
                                                    }
                                                    if (plVar16 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_198;
                                                      plVar8 = local_190;
                                                    }
                                                    else {
                                                      local_1a0 = plVar16;
                                                      iVar4 = FUN_180001510(plVar16);
                                                      if (iVar4 == -1) {
LAB_1800cc298:
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        uVar15 = 0x84;
                                                        plVar9 = local_188;
                                                        plVar7 = local_198;
                                                        plVar8 = local_190;
                                                      }
                                                      else if (iVar4 == 0) {
                                                        iVar4 = (int)*local_168 + 1;
                                                        plVar9 = local_168;
                                                        if (iVar4 != 0) {
                                                          *(int *)local_168 = iVar4;
                                                        }
LAB_1800cc4a9:
                                                        plVar16 = local_168;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#65="storyId" bytes=[3953414, 3953423] */
                                                        uVar12 = DAT_1803b2d08;
                                                        local_198 = plVar9;
                                                        plVar5 = (longlong *)_PyDict_NewPresized(4);
                                                        PyDict_SetItem(plVar5,uVar12,plVar16);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#66="title" bytes=[3953423, 3953430] */
                                                        PyDict_SetItem(plVar5,DAT_1803b2d10,plVar9);
/* Analyst constant binding: novels.pipeline.imitation_short_runner#67="finalMarkdownPath" bytes=[3953430, 3953449] */
                                                        uVar12 = DAT_1803b2d18;
                                                        plVar9 = (longlong *)
/* Analyst constant binding: novels.pipeline.imitation_short_runner#54="final" bytes=[3953308, 3953315] */
                                                                 FUN_180382540(plVar7,DAT_1803b2cb0)
                                                        ;
                                                        if (plVar9 != (longlong *)0x0) {
                                                          plVar8 = (longlong *)
                                                                   FUN_180382540(plVar9,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#55="full.md" bytes=[3953315, 3953324] */
                                                  DAT_1803b2cb8);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar14 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar14 + 0x30))(plVar9);
                                                  }
                                                  if (plVar8 != (longlong *)0x0) {
                                                    plVar9 = (longlong *)FUN_1803749a0(plVar8);
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar14 + 0x30))(plVar8);
                                                    }
                                                    if (plVar9 != (longlong *)0x0) {
                                                      PyDict_SetItem(plVar5,uVar12,plVar9);
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar14 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar14 + 0x30))(plVar9);
                                                      }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#68="resumed" bytes=[3953449, 3953458] */
                                                      PyDict_SetItem(plVar5,DAT_1803b2d20,
                                                                     _Py_TrueStruct_exref);
                                                      plVar9 = local_188;
LAB_1800ccd98:
                                                      FUN_180001d80(param_1);
                                                      if ((-1 < (int)*local_f8) &&
                                                         (*local_f8 = *local_f8 + -1, *local_f8 == 0
                                                         )) {
                                                        lVar14 = FUN_18000cc40(local_130);
                                                        (**(code **)(lVar14 + 0x30))(local_130);
                                                      }
                                                      if ((-1 < (int)*local_e0) &&
                                                         (*local_e0 = *local_e0 + -1, *local_e0 == 0
                                                         )) {
                                                        lVar14 = FUN_18000cc40(local_128);
                                                        (**(code **)(lVar14 + 0x30))(local_128);
                                                      }
                                                      if ((-1 < (int)*local_d8) &&
                                                         (*local_d8 = *local_d8 + -1, *local_d8 == 0
                                                         )) {
                                                        lVar14 = FUN_18000cc40(local_120);
                                                        (**(code **)(lVar14 + 0x30))(local_120);
                                                      }
                                                      if ((-1 < (int)*local_d0) &&
                                                         (*local_d0 = *local_d0 + -1, *local_d0 == 0
                                                         )) {
                                                        lVar14 = FUN_18000cc40(local_118);
                                                        (**(code **)(lVar14 + 0x30))(local_118);
                                                      }
                                                      if ((-1 < (int)*local_c8) &&
                                                         (*local_c8 = *local_c8 + -1, *local_c8 == 0
                                                         )) {
                                                        lVar14 = FUN_18000cc40(local_110);
                                                        (**(code **)(lVar14 + 0x30))(local_110);
                                                      }
                                                      if ((-1 < (int)*local_178) &&
                                                         (*local_178 = *local_178 + -1,
                                                         *local_178 == 0)) {
                                                        lVar14 = FUN_18000cc40(local_108);
                                                        (**(code **)(lVar14 + 0x30))(local_108);
                                                      }
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar14 = FUN_18000cc40(local_148);
                                                        (**(code **)(lVar14 + 0x30))(local_148);
                                                      }
                                                      if ((-1 < (int)*plVar7) &&
                                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                        lVar14 = FUN_18000cc40(local_140);
                                                        (**(code **)(lVar14 + 0x30))(local_140);
                                                      }
                                                      plVar8 = local_f0;
                                                      if (((local_f0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_f0)) &&
                                                         (*local_f0 = *local_f0 + -1, *local_f0 == 0
                                                         )) {
                                                        lVar14 = FUN_18000cc40(local_f0);
                                                        (**(code **)(lVar14 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_1a0;
                                                      if (((local_1a0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_1a0)) &&
                                                         (*local_1a0 = *local_1a0 + -1,
                                                         *local_1a0 == 0)) {
                                                        lVar14 = FUN_18000cc40(local_1a0);
                                                        (**(code **)(lVar14 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_198;
                                                      if (((local_198 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_198)) &&
                                                         (*local_198 = *local_198 + -1,
                                                         *local_198 == 0)) {
                                                        lVar14 = FUN_18000cc40(local_198);
                                                        (**(code **)(lVar14 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_138;
                                                      if ((-1 < (int)*local_138) &&
                                                         (*local_138 = *local_138 + -1,
                                                         *local_138 == 0)) {
                                                        lVar14 = FUN_18000cc40(local_138);
                                                        (**(code **)(lVar14 + 0x30))(plVar8);
                                                      }
                                                      if ((int)*plVar9 < 0) {
                                                        return plVar5;
                                                      }
                                                      *plVar9 = *plVar9 + -1;
                                                      if (*plVar9 != 0) {
                                                        return plVar5;
                                                      }
                                                      lVar14 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar14 + 0x30))(plVar9);
                                                      return plVar5;
                                                    }
                                                  }
                                                  }
                                                  uVar15 = 0x88;
                                                  FUN_18000ce30(param_1,&local_1a8);
                                                  plVar9 = local_188;
                                                  plVar7 = local_198;
                                                  plVar16 = local_1a0;
                                                  plVar8 = local_190;
                                                  if ((-1 < (int)*plVar5) &&
                                                     (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                    lVar14 = FUN_18000cc40(plVar5);
                                                    (**(code **)(lVar14 + 0x30))(plVar5);
                                                    plVar9 = local_188;
                                                    plVar7 = local_198;
                                                    plVar16 = local_1a0;
                                                    plVar8 = local_190;
                                                  }
                                                  }
                                                  else {
                                                    plVar9 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar16,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#62="splitlines" bytes=[3953388, 3953400] */
                                                                           DAT_1803b2cf0);
                                                    plVar8 = local_190;
                                                    if (plVar9 == (longlong *)0x0)
                                                    goto LAB_1800cc298;
                                                    uVar15 = 0x84;
                                                    *(undefined4 *)(local_190 + 5) = 0x84;
                                                    plVar16 = (longlong *)
                                                              FUN_180376210(param_1,plVar9);
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar9);
                                                      (**(code **)(lVar14 + 0x30))(plVar9);
                                                    }
                                                    if (plVar16 == (longlong *)0x0) {
LAB_1800cc318:
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_1a0;
                                                    }
                                                    else {
                                                      plVar9 = (longlong *)
                                                               FUN_1800c6da0(param_1,plVar16,
                                                                             DAT_1803ab250,0);
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar14 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar14 + 0x30))(plVar16);
                                                      }
                                                      if (plVar9 != (longlong *)0x0) {
                                                        plVar16 = (longlong *)
                                                                  FUN_18037f0c0(param_1,plVar9,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#63="lstrip" bytes=[3953400, 3953408] */
                                                                                DAT_1803b2cf8);
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar14 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar14 + 0x30))(plVar9);
                                                        }
                                                        plVar8 = local_190;
                                                        if (plVar16 != (longlong *)0x0) {
                                                          *(undefined4 *)(local_190 + 5) = 0x84;
                                                          plVar9 = (longlong *)
                                                                   FUN_1803770d0(param_1,plVar16,
                                                                                 DAT_1803b2d00);
                                                          if ((-1 < (int)*plVar16) &&
                                                             (*plVar16 = *plVar16 + -1,
                                                             *plVar16 == 0)) {
                                                            lVar14 = FUN_18000cc40(plVar16);
                                                            (**(code **)(lVar14 + 0x30))(plVar16);
                                                          }
                                                          if (plVar9 != (longlong *)0x0) {
                                                            plVar16 = (longlong *)
                                                                      FUN_18037f0c0(param_1,plVar9,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#3="strip" bytes=[3952481, 3952488] */
                                                                                    DAT_1803b2b18);
                                                            if ((-1 < (int)*plVar9) &&
                                                               (*plVar9 = *plVar9 + -1, *plVar9 == 0
                                                               )) {
                                                              lVar14 = FUN_18000cc40(plVar9);
                                                              (**(code **)(lVar14 + 0x30))(plVar9);
                                                            }
                                                            if (plVar16 != (longlong *)0x0) {
                                                              *(undefined4 *)(plVar8 + 5) = 0x84;
                                                              plVar9 = (longlong *)
                                                                       FUN_180376210(param_1,plVar16
                                                                                    );
                                                              if ((-1 < (int)*plVar16) &&
                                                                 (*plVar16 = *plVar16 + -1,
                                                                 *plVar16 == 0)) {
                                                                lVar14 = FUN_18000cc40(plVar16);
                                                                (**(code **)(lVar14 + 0x30))
                                                                          (plVar16);
                                                              }
                                                              if (plVar9 == (longlong *)0x0) {
                                                                FUN_18000ce30(param_1,&local_1a8);
                                                                plVar9 = local_188;
                                                                plVar7 = local_198;
                                                                plVar16 = local_1a0;
                                                                plVar8 = local_190;
                                                                goto LAB_1800cd021;
                                                              }
                                                              goto LAB_1800cc4a9;
                                                            }
                                                          }
                                                          goto LAB_1800cc318;
                                                        }
                                                      }
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_1a0;
                                                      plVar8 = local_190;
                                                    }
                                                  }
                                                  }
                                                  goto LAB_1800cd021;
                                                  }
                                                  }
                                                  FUN_18000ce30(param_1,&local_1a8);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    lVar14 = FUN_18000cc40(plVar9);
                                                    (**(code **)(lVar14 + 0x30))(plVar9);
                                                  }
                                                  uVar15 = 0x83;
                                                  plVar9 = local_188;
                                                  plVar7 = local_198;
                                                  plVar16 = local_198;
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    plVar9 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar8,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                                           DAT_1803b2b00);
                                                    plVar8 = local_190;
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                      uVar15 = 0x82;
                                                      plVar9 = local_188;
                                                      plVar7 = local_198;
                                                      plVar16 = local_198;
                                                      plVar8 = local_190;
                                                    }
                                                    else {
                                                      uVar15 = 0x82;
                                                      *(undefined4 *)(local_190 + 5) = 0x82;
                                                      plVar16 = (longlong *)
                                                                FUN_1803770d0(param_1,plVar9,
                                                                              DAT_1803b2cd8);
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar14 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar14 + 0x30))(plVar9);
                                                      }
                                                      if (plVar16 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        plVar9 = local_188;
                                                        plVar7 = local_198;
                                                        plVar16 = local_198;
                                                      }
                                                      else {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#60="completed" bytes=[3953366, 3953377] */
                                                        iVar4 = FUN_1803836d0(plVar16,DAT_1803b2ce0)
                                                        ;
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar14 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar14 + 0x30))(plVar16);
                                                        }
                                                        if (iVar4 != -1) goto LAB_1800cc11d;
                                                        FUN_18000ce30(param_1,&local_1a8);
                                                        plVar9 = local_188;
                                                        plVar7 = local_198;
                                                        plVar16 = local_198;
                                                      }
                                                    }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      lVar14 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar14 + 0x30))(plVar6);
                                                    }
                                                    plVar5 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar9,
/* Analyst constant binding: novels.pipeline.imitation_short_runner#0="get" bytes=[3952464, 3952469] */
                                                                           DAT_1803b2b00);
                                                    if (plVar5 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                    }
                                                    else {
                                                      *(undefined4 *)(plVar8 + 5) = 0x73;
                                                      plVar6 = (longlong *)
                                                               FUN_1803770d0(param_1,plVar5,
                                                                             DAT_1803b2c70);
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar14 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar14 + 0x30))(plVar5);
                                                      }
                                                      if (plVar6 != (longlong *)0x0) {
                                                        iVar4 = FUN_180001510(plVar6);
                                                        if (iVar4 == -1) goto LAB_1800cb4c9;
                                                        if (iVar4 != 1) {
                                                          if ((-1 < (int)*plVar6) &&
                                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                                          {
                                                            lVar14 = FUN_18000cc40(plVar6);
                                                            (**(code **)(lVar14 + 0x30))(plVar6);
                                                          }
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                          plVar6 = DAT_1803b2b10;
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                          iVar4 = (int)*DAT_1803b2b10 + 1;
                                                          if (iVar4 != 0) {
/* Analyst constant binding: novels.pipeline.imitation_short_runner#2="" bytes=[3952480, 3952481] */
                                                            *(int *)DAT_1803b2b10 = iVar4;
                                                          }
                                                        }
                                                        goto LAB_1800cb5d8;
                                                      }
                                                      FUN_18000ce30(param_1,&local_1a8);
                                                    }
                                                  }
                                                  }
                                                  }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                          goto LAB_1800cd021;
                                        }
                                      }
                                    }
                                    goto LAB_1800cb166;
                                  }
                                  local_1a8 = *(longlong *)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                  {
                                    (**(code **)(plVar6[1] + 0x30))(plVar6);
                                    uVar15 = 0x70;
                                    goto LAB_1800cd021;
                                  }
                                }
                                uVar15 = 0x70;
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
LAB_1800cd021:
  piVar1 = *(int **)(local_1a8 + 0x28);
  if (piVar1 == (int *)0x0) {
    lVar14 = FUN_18038e2a0(plVar8,uVar15);
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar8) goto LAB_1800cd063;
    lVar14 = FUN_18038e2a0(plVar8,uVar15);
    *(int **)(lVar14 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
  }
  FUN_180001300(&local_1a8,lVar14);
LAB_1800cd063:
  plVar6 = local_f0;
  local_1b0 = 0;
  local_1c8 = local_f0;
  local_1d0 = local_140;
  local_1d8 = local_148;
  local_1e0 = local_108;
  local_1e8 = local_110;
  local_1f0 = local_118;
  local_1f8 = local_120;
  local_200 = local_128;
  local_208 = local_130;
  local_1c0 = plVar16;
  local_1b8 = plVar7;
  FUN_18036d0a0(plVar8,"oooooooooooooo",local_138,plVar9);
  plVar5 = DAT_1803b3618;
  if (plVar8 == DAT_1803b3618) {
    if ((-1 < (int)*DAT_1803b3618) && (*DAT_1803b3618 = *DAT_1803b3618 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))();
    }
    DAT_1803b3618 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  lVar14 = local_1a8;
  local_1a8 = 0;
  if (((local_130 != (longlong *)0x0) && (-1 < (int)*local_130)) &&
     (*local_130 = *local_130 + -1, *local_130 == 0)) {
    (**(code **)(local_130[1] + 0x30))();
  }
  if (((local_128 != (longlong *)0x0) && (-1 < (int)*local_128)) &&
     (*local_128 = *local_128 + -1, *local_128 == 0)) {
    (**(code **)(local_128[1] + 0x30))();
  }
  if (((local_120 != (longlong *)0x0) && (-1 < (int)*local_120)) &&
     (*local_120 = *local_120 + -1, *local_120 == 0)) {
    (**(code **)(local_120[1] + 0x30))();
  }
  if (((local_118 != (longlong *)0x0) && (-1 < (int)*local_118)) &&
     (*local_118 = *local_118 + -1, *local_118 == 0)) {
    (**(code **)(local_118[1] + 0x30))();
  }
  if (((local_110 != (longlong *)0x0) && (-1 < (int)*local_110)) &&
     (*local_110 = *local_110 + -1, *local_110 == 0)) {
    (**(code **)(local_110[1] + 0x30))();
  }
  if (((local_108 != (longlong *)0x0) && (-1 < (int)*local_108)) &&
     (*local_108 = *local_108 + -1, *local_108 == 0)) {
    (**(code **)(local_108[1] + 0x30))();
  }
  if (((local_148 != (longlong *)0x0) && (-1 < (int)*local_148)) &&
     (*local_148 = *local_148 + -1, *local_148 == 0)) {
    (**(code **)(local_148[1] + 0x30))();
  }
  if (((local_140 != (longlong *)0x0) && (-1 < (int)*local_140)) &&
     (*local_140 = *local_140 + -1, *local_140 == 0)) {
    (**(code **)(local_140[1] + 0x30))();
  }
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))(plVar6);
  }
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  local_1a8 = lVar14;
  if ((-1 < (int)*local_138) && (*local_138 = *local_138 + -1, *local_138 == 0)) {
    (**(code **)(local_138[1] + 0x30))();
  }
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  FUN_18000ce40(param_1,&local_1a8);
  return (longlong *)0x0;
}
